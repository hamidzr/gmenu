# gmenu / gomenu

## TODO
- make lint and fix
- x initial  query not working
- [ ] ctrl+c to work like esc
- [ ] fuzzy search avoid matching single chars alone unless near separator?
- [ ] report created window's id and pid?
- [ ] multiple selection: select two inputs for guillm.
- [ ] no focus border color for input
- [ ] search: fuzzy with char presence check and relative count
- [ ] add cli support for same behavior but in the terminal
- [ ] survey dmenu option to provide better compatibility
  - add prompt option
- [ ] preserve original input order on startup (and resets?)
- [x] auto pick single list of options. preapply a query as well
- [x] remember last entry and have it auto selected on start.
  - or put it as the first option..
- [x] close on focus loss
- [x] instance lock only based on menu name. pid
- app should have a name for activity monitor and ps faux
- [x] a single instance with the same title
- [x] the current fuzzy search is not great
- [x] simple fuzzy to ignore some letters eg space - \_

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
