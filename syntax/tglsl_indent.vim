" Vim indent file
" Language:	wtsuite tglsl (dialect of glsl)
" Filenames: *.tglsl

if exists("b:did_indent")
   finish
endif

let b:did_indent = 1

" C indenting is built-in, thus this is very simple
setlocal cindent

let b:undo_indent = "setl cin<"
