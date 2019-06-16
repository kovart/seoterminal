# Seoterminal

This app helps to work with very large CSV-files of **SEO keywords**.
It's fully written on GoLang so it gives the maximum of possible performance.
The user interface is terminal-based so you won't feel any lags or freezing.

> It was my first GoLang experience. The project is currently not maintained. 

<p align="center">
  <img alt="Demonstration" src="https://github.com/kovart/seoterminal/raw/assets/demo.gif?raw=true"></img>
</p>

## Features 
- Search by keyword
- Grouping by lemmatized words
- Various types of sorting
- Cut groups into separeted files
- Expand nested groups as deep as you want
- History with the ability to restore states

## How to run
To create a new project:
```sh
$ ./tool -p <ProjectName> -f <PathToSourceFile>
```
To load existing project:
```sh
$ ./tool -p <PathToProject>
```

## Hotkeys 
* <kbd>+</kbd> : Save cluster into separeted file
* <kbd>-</kbd> : Save cluster into separeted file in `removed` folder
* <kbd>Ctrl</kbd> + <kbd>A</kbd> : Reset root cluster
* <kbd>Ctrl</kbd> + <kbd>K</kbd> : Set the current cluster as root
* <kbd>Ctrl</kbd> + <kbd>D</kbd> : Remove the current cluster as root
* <kbd>Ctrl</kbd> + <kbd>S</kbd> : Save the current cluster as root
* <kbd>Alt</kbd> + <kbd>H</kbd> and <kbd>Esc</kbd> : Open and Close history
* <kbd>Tab</kbd> and <kbd>Shift</kbd> + <kbd>Tab</kbd> : Nivagation

## How to build
The first you need to install all dependencies:
```sh
$ go get .
```
Then build the runnable file:
```sh
# For windows:
$ go build -o tool.exe
# For Linux/MacOS:
$ go build -o tool
```

## Dependencies
* [tview](https://github.com/rivo/tview)
* [csvutil](https://github.com/jszwec/csvutil)
