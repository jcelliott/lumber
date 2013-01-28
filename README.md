lumber
======

A simple logger for Go.

###Features:###
- Log levels
- Console logger
- File logger
- Log backup
- Log prefixes
- Buffered file logger

Todo:
- Log rotation

###Usage:###
Log to the default (console) logger  
  `lumber.Error("An error message")` etc.

Create a new console logger that only logs messages of level WARN or higher  
  `log := lumber.NewConsoleLogger(lumber.WARN)`
  
Create a new file logger (rotating at 5000 lines, 200 message buffer)  
  `log := lumber.NewFileLogger("filename.log", lumber.DEBUG, 5000, 200)`

Add a prefix to label different logs  
  `log.Prefix("MYAPP")`

And that's all.
