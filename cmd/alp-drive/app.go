package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	usecasefile "alp-drive/internal/application/usecases/file"
	usecasefolder "alp-drive/internal/application/usecases/folder"
	web "alp-drive/internal/delivery/web"
	"alp-drive/internal/infrastructure/localfs"
	"alp-drive/internal/infrastructure/sqlite"
	"alp-drive/internal/infrastructure/system"
)

type Config struct {
	HTTPAddr  string
	SQLiteDSN string
	FilesDir  string
	LogLevel  slog.Level
}

func loadConfig() Config {
	var level slog.Level
	if err := level.UnmarshalText([]byte(mustEnv("ALP_DRIVE_LOG_LEVEL"))); err != nil {
		panic(fmt.Sprintf("invalid ALP_DRIVE_LOG_LEVEL: %v", err))
	}

	return Config{
		HTTPAddr:  mustEnv("ALP_DRIVE_HTTP_ADDR"),
		SQLiteDSN: mustEnv("ALP_DRIVE_DB_PATH"),
		FilesDir:  mustEnv("ALP_DRIVE_FILES_DIR"),
		LogLevel:  level,
	}
}

func mustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic("missing required environment variable: " + key)
	}
	return v
}

type UseCases struct {
	ListFolders  *usecasefolder.ListFoldersUC
	CreateFolder *usecasefolder.CreateFolderUC
	RenameFolder *usecasefolder.RenameFolderUC
	MoveFolder   *usecasefolder.MoveFolderUC
	DeleteFolder *usecasefolder.DeleteFolderUC

	ListFiles  *usecasefile.ListFilesByFolderUC
	CreateFile *usecasefile.CreateFileUC
	RenameFile *usecasefile.RenameFileUC
	MoveFile   *usecasefile.MoveFileUC
	DeleteFile *usecasefile.DeleteFileUC
	Download   *usecasefile.DownloadFileUC
}

type App struct {
	cfg        Config
	useCases   UseCases
	HTTPServer *http.Server
	closers    []func() error
}

func NewApp() (*App, error) {
	cfg := loadConfig()
	ctx := context.Background()

	db, err := sqlite.Open(ctx, cfg.SQLiteDSN)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	closers := []func() error{db.Close}

	store := sqlite.NewStore(db)
	files := localfs.NewFileStore(cfg.FilesDir)
	clk := system.NewClock()

	useCases := UseCases{
		ListFolders:  usecasefolder.NewListFoldersUC(store),
		CreateFolder: usecasefolder.NewCreateFolderUC(store, clk),
		RenameFolder: usecasefolder.NewRenameFolderUC(store, clk),
		MoveFolder:   usecasefolder.NewMoveFolderUC(store, clk),
		DeleteFolder: usecasefolder.NewDeleteFolderUC(store, files),

		ListFiles:  usecasefile.NewListFilesByFolderUC(store),
		CreateFile: usecasefile.NewCreateFileUC(store, files, clk),
		RenameFile: usecasefile.NewRenameFileUC(store, clk),
		MoveFile:   usecasefile.NewMoveFileUC(store, clk),
		DeleteFile: usecasefile.NewDeleteFileUC(store, files),
		Download:   usecasefile.NewDownloadFileUC(store, files),
	}

	logger := newLogger(cfg)
	handler := web.NewDriveHandler(
		useCases.ListFolders,
		useCases.CreateFolder,
		useCases.RenameFolder,
		useCases.MoveFolder,
		useCases.DeleteFolder,
		useCases.ListFiles,
		useCases.CreateFile,
		useCases.RenameFile,
		useCases.MoveFile,
		useCases.DeleteFile,
		useCases.Download,
		logger,
	)

	return &App{
		cfg:      cfg,
		useCases: useCases,
		HTTPServer: &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: web.NewRouter(handler),
		},
		closers: closers,
	}, nil
}

func (a *App) Close() {
	closeAll(a.closers)
}

func closeAll(closers []func() error) {
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i](); err != nil {
			fmt.Fprintf(os.Stderr, "closing resource: %v\n", err)
		}
	}
}

func newLogger(cfg Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel}))
}
