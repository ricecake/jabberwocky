package util

/***** TODO:
This should hold some functions relating to setting up some more universal logging.

Should be able to grab the io.Writter from the standard logging package, and replace it with one that diverts the log leval and message into apex logs.
Should then be able to make the apex log have a cli log handler that will redirect to the original io.Writer, as well as any other configured loggers.
That way it can be configured what loggers to enable, including forwarding to other servers, if that's needed.
Should be able to work on both the client and the server.
*/
