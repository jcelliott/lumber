lumber
======

A simple logger for Go.

###Features:###
- Log levels
- Console logger
- File logger
- Log backup
- Log prefixes

Todo:
- Log rotation
- Buffered file logger

###Usage:###
Log to the default (console) logger  
  `lumber.Error("An error message")` etc.

Create a new console logger that only logs messages of level INFO or higher  
  `log := lumber.NewConsoleLogger(lumber.INFO)`
  
Create a new file logger (rotating at 1000 lines)  
  `log := lumber.NewFileLogger("filename.log", lumber.DEBUG, 1000)`

Add a prefix to label different logs  
  `log.Prefix("APP")`

And that's all.
