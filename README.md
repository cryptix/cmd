livefriday
==========

[blackfriday](https://github.com/russross/blackfriday)+[livereload](https://github.com/jaschaephraim/lrserver) - a simple markdown to html generator wich livereloads on save

## TODO
- [ ] Pass `watchDir` as an argument
- [ ] Currently only lists and reloads files with `.md` suffix
- [ ] Configure http server host and port
- [x] Open browser automatically after launch - using [open](https://github.com/skratchdot/open-golang)
- [x] Check error from `http.ListenAndServer()`
- [x] **!security issue!** `mdHandler` loads files outside `watchDir` - now using `filepath.Base()` to stop local file inclusions
