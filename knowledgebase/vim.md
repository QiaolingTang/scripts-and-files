# basic commands
`c` change

`d` delete

`y` copy

`p` paste after the cursor

`P` paste before the cursor

`r` replace

`yy` copy line

`dd` delete line


# move
`0` move to the beginning of a line

`^` go to the first non-blank character in a line

`$` move to the end of a line

`gg` move to the beginning of a file

`G` move to the end of a file

`ctrl f` move page forward

`ctrl b` move page backward

`)` move to the beginning of next sentence

`(` move to the previous sentence

`}` move to the next paragraph

`{` move to the previous paragraph

`/` search forwards

`?` search backwards

`ctrl o` move backward to the jump list

`ctrl l` move forward to the jump list

`:jumps` list jump list

`'.` move to last change


# windows and buffers
`:split` split windows into same buffer

`ctrl w w` move between windows

`ctrl w c` close window

`:new`+$filename open another file in same window

`:e`/`:edit` edit new file or another file

`:ls` list buffers

`:b`+number move to the specific buffer(number is found by `:ls`)

`:bd`+$filename delete buffer

`:e`+tab list buffers

`:e`+$prefix+tab list files start with the prefix

`:bn`, `:bp` move between buffers

`:b`+space+$filename move to specific buffer

`:e!` load the latest version of the file and disregard the changes



# mode
- `shift v` visual line

- `ctrl v` visual block


# registers
`"`+ a-zA-Z, e.g.:
- `"a yy` yank the line to the register a
- `"a p` paste from register a


# change until character
- `cf`+character `f` won't highlight the searched character.

- `ct`+character the character is not included.


# change block:
1. enter visual mode
2.  select block
3.  `c`
4.  enter new charactors
5.  go back to normal mode

# replace
`s/old/new` only replace the first occurrence

`s/old/new/g` for global


# marks
`m` create mark, e.g.:
1.  go to the exact line
2.  create a mark: `ma`, notes: `m`+a-zA-Z set the mark name, e.g.: `ma`, `a-z` local to the current buffer, `A-Z` global marks
3.  `'a` back to the mark


# read
`:r`+$filename read $filename into current file


# open file
`gf`/`gF` go to file, e.g.:

1. open a file in vim with below content:
```
FROM fedora:latest as builder

RUN sudo yum install -y go

COPY multiline-log.go run-go.sh /
RUN go build /multiline-log.go && mkdir -p /var/lib/logging/ && chmod +x ./multiline-log
COPY multiline-log.cfg /var/lib/logging/
WORKDIR /
```
2. move cursor to `run-go.sh`
3. press `g`+`f`/`F`


# run external command in vim
`:!`+cmd

`:%! jq .` run `jq` command in current file, `%` means current file

`:r !`+cmd read the output of cmd and put it into current file

# run one Normal mode command in Insert mode
`Ctrl o` + Normal mode command

# stop vim from adding a new line at the end of a file
```
vim -b $filename
:set noeol
:wq
```

# useful links
- https://vimgolf.com/
- https://vim-adventures.com/
- https://vim.rtorr.com/
