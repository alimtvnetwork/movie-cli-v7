// movie_scan.go — movie scan [folder] — command definition and orchestrator
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alimtvnetwork/movie-cli-v5/cleaner"
	"github.com/alimtvnetwork/movie-cli-v5/db"
	"github.com/alimtvnetwork/movie-cli-v5/errlog"
	"github.com/alimtvnetwork/movie-cli-v5/tmdb"
)

var scanRecursive bool
var scanDepth int
var scanDryRun bool
var scanFormat string
var scanRest bool
var scanRestPort int

var movieScanCmd = &cobra.Command{
	Use:   "scan [folder]",
	Short: "Scan a folder for movies and TV shows",
	Long: `Scans a folder for video files, cleans filenames, fetches metadata
from TMDb, downloads thumbnails, and stores everything in the database.

If no folder is specified, scans the current working directory.
Use --recursive (-r) to scan all subdirectories recursively.
Use --depth to limit how many levels deep the recursive scan goes.
Use --dry-run to preview what would be scanned without writing anything.

Examples:
  movie scan                      Scan current directory (top-level)
  movie scan ~/Movies             Scan specific folder
  movie scan -r                   Scan current directory recursively
  movie scan --dry-run            Preview files without writing to DB
  movie scan --format table       Show results as a formatted table
  movie scan --rest               Scan and start REST server + open browser`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMovieScan,
}

func init() {
	movieScanCmd.Flags().BoolVarP(&scanRecursive, "recursive", "r", false,
		"scan all subdirectories recursively")
	movieScanCmd.Flags().IntVarP(&scanDepth, "depth", "d", 0,
		"max subdirectory depth for recursive scan (0 = unlimited)")
	movieScanCmd.Flags().BoolVar(&scanDryRun, "dry-run", false,
		"preview what would be scanned without writing to DB or .movie-output")
	movieScanCmd.Flags().StringVar(&scanFormat, "format", "default",
		"output format: default, table, or json")
	movieScanCmd.Flags().BoolVar(&scanRest, "rest", false,
		"start REST server and open HTML report in browser after scan")
	movieScanCmd.Flags().IntVar(&scanRestPort, "port", 8086,
		"port for REST server when using --rest")
	movieScanCmd.Flags().BoolVarP(&scanWatch, "watch", "w", false,
		"watch for new files after initial scan")
	movieScanCmd.Flags().IntVar(&scanWatchInterval, "interval", 10,
		"polling interval in seconds for --watch mode")
}

func runMovieScan(cmd *cobra.Command, args []string) {
	useJSON := scanFormat == "json"

	database, err := db.Open()
	if err != nil {
		errlog.Error(msgDatabaseError, err)
		return
	}
	defer database.Close()

	scanDir, err := resolveScanDir(args, useJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		return
	}

	creds := resolveScanTMDbCredentials(database)
	outputDir := filepath.Join(scanDir, ".movie-output")

	if !scanDryRun {
		if err := createOutputDirs(outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			return
		}
	}

	initScanLogger(database, outputDir)

	if !useJSON {
		printScanHeader(scanDir, outputDir)
	}

	ctx := createScanContext(database, creds, outputDir)
	removed, jsonItems := executeScan(ctx, scanDir, useJSON)
	finalizeScan(cmd, ctx, FinalizeScanInput{
		ScanDir: scanDir, OutputDir: outputDir, Database: database,
		Creds: creds, Removed: removed, JSONItems: jsonItems, UseJSON: useJSON,
	})
}

func createScanContext(database *db.DB, creds tmdbCredentials, outputDir string) *ScanContext {
	tmdbClient := tmdb.NewClientWithToken(creds.APIKey, creds.Token)
	tmdbClient.SetIMDbCache(newIMDbCacheAdapter(database))
	return &ScanContext{
		Database:  database,
		Client:    tmdbClient,
		HasTMDb:   creds.HasAuth(),
		OutputDir: outputDir,
		UseTable:  scanFormat == string(db.OutputFormatTable) || scanFormat == "json",
		BatchID:   generateBatchID(),
	}
}

func executeScan(ctx *ScanContext, scanDir string, useJSON bool) (int, []scanJSONItem) {
	videoFiles := collectVideoFiles(scanDir, scanRecursive, scanDepth)
	useTable := scanFormat == string(db.OutputFormatTable)
	var jsonItems []scanJSONItem

	if useTable {
		printScanTableHeader()
	}

	if scanDryRun {
		runDryRunScan(DryRunInput{VideoFiles: videoFiles, UseJSON: useJSON, UseTable: useTable},
			DryRunOutput{JSONItems: &jsonItems, TotalFiles: &ctx.TotalFiles, MovieCount: &ctx.MovieCount, TVCount: &ctx.TVCount})
		if useTable {
			printScanTableFooter()
		}
		return 0, jsonItems
	}

	removed := runMainScanLoop(ctx, videoFiles, ScanLoopConfig{
		Client: ctx.Client, ScanDir: scanDir, BatchID: ctx.BatchID,
		UseJSON: useJSON, UseTable: useTable, HasTMDb: ctx.HasTMDb,
		JSONItems: &jsonItems,
	})

	if useTable {
		printScanTableFooter()
	}
	return removed, jsonItems
}

func finalizeScan(cmd *cobra.Command, ctx *ScanContext, input FinalizeScanInput) {
	registerScanHistory(input.Database, input.ScanDir, ctx)

	if input.UseJSON {
		printScanJSON(input.ScanDir, input.JSONItems, ScanStats{
			Total: ctx.TotalFiles, Movies: ctx.MovieCount, TV: ctx.TVCount, Skipped: ctx.Skipped,
		})
	}
	if !input.UseJSON {
		printScanFooter(ScanStats{
			ScanDir: input.ScanDir, OutputDir: input.OutputDir, Items: ctx.ScannedItems,
			Total: ctx.TotalFiles, Movies: ctx.MovieCount, TV: ctx.TVCount,
			Skipped: ctx.Skipped, Removed: input.Removed,
		})
	}

	tmdbClient := tmdb.NewClientWithToken(input.Creds.APIKey, input.Creds.Token)
	tmdbClient.SetIMDbCache(newIMDbCacheAdapter(input.Database))
	startPostScanServices(cmd, ScanServiceConfig{
		ScanDir: input.ScanDir, OutputDir: input.OutputDir, Database: input.Database, Creds: input.Creds,
	}, tmdbClient)
}

func initScanLogger(database *db.DB, outputDir string) {
	if initErr := errlog.Init(outputDir, "scan"); initErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not init error logger: %v\n", initErr)
		return
	}
	errlog.SetDBWriter(func(e errlog.Entry) {
		dbErr := database.InsertErrorLog(db.ErrorLogEntry{
			Timestamp: e.Timestamp, Level: string(e.Level), Source: e.Source,
			Function: e.Function, Command: e.Command, WorkDir: e.WorkDir,
			Message: e.Message, StackTrace: e.StackTrace,
		})
		if dbErr != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Could not write error to DB: %v\n", dbErr)
		}
	})
}

func registerScanHistory(database *db.DB, scanDir string, ctx *ScanContext) {
	if scanDryRun {
		return
	}
	folderId, folderErr := database.UpsertScanFolder(scanDir)
	if folderErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not register scan folder: %v\n", folderErr)
		return
	}
	if histErr := database.InsertScanHistory(db.ScanHistoryInput{
		ScanFolderID: int(folderId), TotalFiles: ctx.TotalFiles,
		Movies: ctx.MovieCount, TV: ctx.TVCount,
	}); histErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not log scan history: %v\n", histErr)
	}
}

func startPostScanServices(cmd *cobra.Command, cfg ScanServiceConfig, client *tmdb.Client) {
	if scanDryRun {
		return
	}
	if scanRest {
		startRestWithOptionalWatch(cmd, cfg)
		return
	}
	if scanWatch {
		runWatchLoop(cfg)
	}
}

func startRestWithOptionalWatch(cmd *cobra.Command, cfg ScanServiceConfig) {
	restPort = scanRestPort
	fmt.Printf("\n🚀 Starting REST server on http://localhost:%d ...\n", restPort)
	go openBrowser(fmt.Sprintf("http://localhost:%d", restPort))
	if scanWatch {
		go runWatchLoop(cfg)
	}
	runMovieRest(cmd, []string{})
}

// runDryRunScan handles the dry-run scanning loop for all output formats.
func runDryRunScan(input DryRunInput, output DryRunOutput) {
	if input.UseJSON {
		items, mc, tc := buildDryRunJSONItems(input.VideoFiles)
		*output.JSONItems = items
		*output.TotalFiles = len(items)
		*output.MovieCount = mc
		*output.TVCount = tc
		return
	}
	if input.UseTable {
		rows, mc, tc := buildDryRunTableRows(input.VideoFiles)
		for _, row := range rows {
			printScanTableRow(row)
		}
		*output.TotalFiles = len(rows)
		*output.MovieCount = mc
		*output.TVCount = tc
		return
	}
	runDryRunPlainOutput(input.VideoFiles, DryRunCounters{
		TotalFiles: output.TotalFiles, MovieCount: output.MovieCount, TVCount: output.TVCount,
	})
}

// runDryRunPlainOutput prints dry-run results in plain text format.
func runDryRunPlainOutput(videoFiles []videoFile, counters DryRunCounters) {
	for _, vf := range videoFiles {
		*counters.TotalFiles++
		result := cleaner.Clean(vf.Name)
		typeIcon := db.TypeIcon(result.Type)
		fmt.Printf("\n  %d. %s %s", *counters.TotalFiles, typeIcon, result.CleanTitle)
		if result.Year > 0 {
			fmt.Printf(" (%d)", result.Year)
		}
		fmt.Printf(" [%s]\n", result.Type)
		fmt.Printf("     └─ %s\n", vf.Name)
		incrementTypeCountPtr(result.Type, counters.MovieCount, counters.TVCount)
	}
}

// incrementTypeCountPtr bumps movie or tv count pointer based on media type.
func incrementTypeCountPtr(mediaType string, movieCount, tvCount *int) {
	if mediaType == string(db.MediaTypeMovie) {
		*movieCount++
		return
	}
	*tvCount++
}
