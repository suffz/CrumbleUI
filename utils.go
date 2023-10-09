package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//

var D = []string{}

func Len() int {
	return len(D)
}

func Canvas() fyne.CanvasObject {
	return &widget.Label{Text: "template", TextStyle: widget.RichTextStyleCodeBlock.TextStyle}
}

func Update(i widget.ListItemID, o fyne.CanvasObject) {
	o.(*widget.Label).SetText(D[i])
}

func ResizeAndShowDialog(Dialog dialog.Dialog) {
	Dialog.Resize(fyne.NewSize(400, 120))
	Dialog.Show()
}

func AddMessage(Content string, list *widget.List) {
	D = append(D, Content)
	list.Refresh()
	list.ScrollToBottom()
}

func TempCalc(interval, accamt int) time.Duration {
	return time.Duration(interval/accamt) * time.Millisecond
}
