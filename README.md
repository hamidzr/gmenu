# gmenu / gomenu

## TODO

- [x] remember last entry and have it auto selected on start.
  - or put it as the first option..
- [x] close on focus loss
- [ ] instance lock only based on menu name. pid
- app should have a name for activity monitor and ps faux
- [x] a single instance with the same title
- [x] the current fuzzy search is not great
- add cli support for same behavior but in the terminal
- survey dmenu option to provide compatibility
  - add prompt option
- preserve original input order on startup (and resets?)
- [x] simple fuzzy to ignore some letters eg space - \_
- [ ] search: fuzzy with char presense check and relative count

```
choose:
 -i           return index of selected element
 -v           show choose version
 -n [10]      set number of rows
 -w [50]      set width of choose window
 -s [26]      set font size used by choose
 -c [0000FF]  highlight color for matched string
 -b [222222]  background color of selected element
 -m           return the query string in case it doesn't match any item
 -p           defines a prompt to be displayed when query field is empty
 -o           given a query, outputs results to standard output
```

## Inspiration

- suckless dmenu
- dmenu-mac
- rofi
