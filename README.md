lumber
======

A simple logger for Go.

### Features: ###
- Log levels
- Console logger
- File logger
- Log backup
- Log prefixes

Todo:
- Log rotation

### Usage: ###
Log to the default (console) logger  
  `lumber.Error("An error message")` etc.

Create a new console logger that only logs messages of level INFO or higher  
  `log := lumber.NewConsoleLogger(lumber.INFO)`
  
Create a new file logger (rotating at 1000 lines)  
  `log := lumber.NewFileLogger("filename.log", lumber.DEBUG, 1000)`

Add a prefix to label different logs  
  `log.Prefix("APP")`

And that's all.

### Modes: ###

APPEND: Append if the file exists, otherwise create a new file

TRUNC: Open and truncate the file, regardless of whether it already exists

BACKUP: Rotate the log every time a new logger is created

ROTATE: Append if the file exists, when the log reaches maxLines rotate files
