# lists of all the htmlpp command-line tools 
cmds = wt-site wt-search-index wt-template wt-template-syntax-tree wt-script wt-script-syntax-tree wt-svg-minify wt-script-refactor wt-script-grapher wt-glsl wt-glsl-syntax-tree wt-pkg-sync wt-style wt-crawl wt-serve wt-json

version = 0.6.0

# temporary destination directory of the commands, also used by temporary files from math-font
export build = $(abspath ./build)

build_windows_amd64=$(abspath ./build/windows_amd64)
build_darwin_amd64=$(abspath ./build/darwin_amd64)

# installation directory of the commands
prefix = /usr/local/bin

.PHONY: math-font
# package files on which all the commands depend
pkg = $(shell find ./pkg/ -name \*.go)

GIT_COMMIT=$(shell git rev-list -1 HEAD)

dsts = $(addprefix $(build)/,$(cmds))

dsts_windows_amd64 = $(addprefix $(build_windows_amd64)/,$(cmds))
dsts_darwin_amd64 = $(addprefix $(build_darwin_amd64)/,$(cmds))

all: math-font $(dsts)

alt: math-font $(dsts_windows_amd64) $(dsts_darwin_amd64)

math-font:
	make -C $@

build_flags=-ldflags "-X main.VERSION=$(version)"

.SECONDEXPANSION:

$(dsts): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build)
	cd $(dir $<); \
	go build -o $(abspath $@) $(build_flags)

# TODO: can we do better than just copying these rules for different os's/architectures?
$(dsts_windows_amd64): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build_windows_amd64)
	export GOOS=windows; \
	export GOARCH=amd64; \
	cd $(dir $<); \
	go build -o $(abspath $@) $(build_flags)
	
# TODO: can we do better than just copying these rules for different os's/architectures?
$(dsts_darwin_amd64): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build_darwin_amd64)
	export GOOS=darwin; \
	export GOARCH=amd64; \
	cd $(dir $<); \
	go build -o $(abspath $@) $(build_flags)

$(build) $(build_windows_amd64) $(build_darwin_amd64):
	mkdir -p $@

clean:
	rm $(build)/*

install: all
	sudo cp -t $(prefix) $(dsts)

install-vim: install-vim-syntax install-vim-indent

install-vim-%:
ifneq ($(wildcard $(HOME)/.vim/$*),)
	cp ./syntax/thtml_$*.vim $(HOME)/.vim/$*/thtml.vim
	cp ./syntax/tjs_$*.vim $(HOME)/.vim/$*/tjs.vim
	cp ./syntax/tglsl_$*.vim $(HOME)/.vim/$*/tglsl.vim
else ifneq ($(wildcard $(HOME)/.config/nvim/$*),)
	cp ./syntax/thtml_$*.vim $(HOME)/.config/nvim/$*/thtml.vim
	cp ./syntax/tjs_$*.vim $(HOME)/.config/nvim/$*/tjs.vim
	cp ./syntax/tglsl_$*.vim $(HOME)/.config/nvim/$*/tglsl.vim
else
	$(warning no vim $* directory found)
endif

uninstall:
	sudo rm -f $(addprefix $(prefix)/,$(cmds))

package: 
	tar -czf ./build/wtsuite_$(version)-linux_x86_64.tar.gz ./build/wt-*;
	tar -czf ./build/wtsuite_$(version)-mac_darwin.tar.gz $(build_darwin_amd64)/wt-*;
	zip ./build/wtsuite_$(version)-ms_windows.zip $(build_windows_amd64)/wt-*;
	tar -czf ./build/wtsuite_$(version)-vim_syntax.tar.gz ./syntax/*.vim;

release: all alt package
