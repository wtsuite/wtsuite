" Language: wtsuite tjs (dialect of javascript)
" Filenames: *.tjs

if exists("b:did_indent")
  finish
endif

" identical to javascript
runtime! indent/javascript.vim
