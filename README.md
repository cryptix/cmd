livefriday
==========

[blackfriday](https://github.com/russross/blackfriday)+[livereload](https://github.com/jaschaephraim/lrserver) - a simple markdown to html generator wich livereloads on save

## TODO
- [ ] Pass `watchDir` as an argument
- [ ] Currently only lists and reloads files with `.md` suffix
- [ ] Configure http server host and port
- [ ] Open browser automatically after launch
- [x] **!security issue!** `mdHandler` loads files outside `watchDir` - now using `filepath.Base()` to stop local file inclusions
