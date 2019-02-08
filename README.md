# LED
LED is a lightweight editor written using Go.

## Disclaimer
This project is in VERY early stages and you could very likely destroy any file you open with it. _SO USE WITH CAUTION!_

## Installation
    git clone https://github.com/leothelocust/led.git
    cd led/
    make testfile
    make build
    ./editor tmp.txt

## Supported Key Bindings

|Key Binding|Action|
|---|---|
| **Main** | |
|`C-x C-s` | Save |
|`C-x C-c` | Quit |
|`C-g` | Abort current key command |
| **Movement** | |
|`C-p`\|`UP ARROW` | Up |
|`C-n`\|`DOWN ARROW` | Down |
|`C-f`\|`RIGHT ARROW` | Right |
|`C-b`\|`LEFT ARROW` | Left |
|`C-e` | End of line |
|`C-a` | Beginning of line |
|`M-b` | Move Back by Word |
|`M-f` | Move Forward by Word |
| **Other** | |
|`C-u` | Prefix multiplier (similar to emacs) |
| **Destructive** | |
|`Backspace`| ... this is obvious |
|`C-k`| Kill from point forward |
| **Insert** | |
|`[a-zA-Z0-9]`| Insert Character at point |
|`!@#$%^&*()-=;\'"/.,`| Insert Special Character at point |



## TODO
* support more actions
  * like delete (not just backspace)
* undo
* redo

## Known Issues
_... to many to list_

## Contribute
Not actively looking for help at this point, as this project is _very_ early in development.  That being said, if you see something that is fundamentally wrong, please don't just sit there, let me know.
