package tui

import "fmt"

// Need to support all the different color types (cf ansi escape seqs)
// terminals can have great support
// func GetTermFgColor(colorName string) string {
//   switch colorName {
//   case "DEFAULT":
//     return "\x1b[39m"
//   case "BLACK":
//
//
//     case "BLACK":
//   }
// }
//
// func GetTermBgColor(colorName string) string {
//
// }

// TODO implement this

func GetTermColor(colorId int) string {
  return fmt.Sprintf("\x1b[%dm", colorId)
}

func GetFgStyleReset() string {
  return "\x1b[39m"
}

func GetBgStyleReset() string {
  return "\x1b[49m"
}

func GetStyleReset() string {
  return GetFgStyleReset()+GetBgStyleReset()
}

func Get256FgTermColor(colorId int) string {
  return fmt.Sprintf("\x1b[38;5;%dm",colorId)
}

func Get256BgTermColor(colorId int) string {
  return fmt.Sprintf("\x1b[48;5;%dm",colorId)
}

