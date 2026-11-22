## My Fork

This is a personal fork of [mygo-music](https://github.com/egoist/mygo-music), with additional fixes and improvements.

### Changes

- Fixed local LRC lyrics file matching.
- Improved support for LRC files with line start and end timestamps.
- Added support for word-level timestamps in bracketed LRC formats.
- Improved parsing of timed lyrics.
- Added tests for local lyrics matching and parsing.

### Building on Windows

To build the application without displaying a console window:

```powershell
go build -ldflags="-H windowsgui" -o MyGoMusic.exe .
```

To embed a custom application icon, first generate a Windows resource file:

```powershell
rsrc -ico icon.ico -o rsrc_windows_amd64.syso
```

Then build the application:

```powershell
go build -ldflags="-H windowsgui" -o MyGoMusic.exe .
```
