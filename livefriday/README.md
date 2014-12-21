livefriday
==========

[blackfriday](https://github.com/russross/blackfriday)+[livereload](https://github.com/jaschaephraim/lrserver) - a simple markdown to html generator wich livereloads on save

## TODO
- [x] Pass `watchDir` as an argument
- [x] Currently only lists and reloads files with `.md` suffix - added `.mdown`
- [x] Configure http server host and port - using [codegangsta/cli](https://github.com/codegangsta/cli)
- [x] Open browser automatically after launch - using [open](https://github.com/skratchdot/open-golang)
- [x] Check error from `http.ListenAndServer()`
- [x] **!security issue!** `mdHandler` loads files outside `watchDir` - now using `filepath.Base()` to stop local file inclusions
