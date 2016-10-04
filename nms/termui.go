package nms

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui"
)

// Widgets represents termui widgets
type Widgets struct {
	header   *ui.Par
	menu     *ui.Par
	ifName   *ui.List
	ifStatus *ui.List
	ifDescr  *ui.List
	ifTIn    *ui.List
	ifTOut   *ui.List
	ifPIn    *ui.List
	ifPOut   *ui.List
	ifDIn    *ui.List
	ifDOut   *ui.List
	ifEIn    *ui.List
	ifEOut   *ui.List
}

func initWidgets() *Widgets {
	return &Widgets{
		header:   ui.NewPar(""),
		menu:     ui.NewPar(""),
		ifName:   ui.NewList(),
		ifStatus: ui.NewList(),
		ifDescr:  ui.NewList(),
		ifTIn:    ui.NewList(),
		ifTOut:   ui.NewList(),
		ifPIn:    ui.NewList(),
		ifPOut:   ui.NewList(),
		ifDIn:    ui.NewList(),
		ifDOut:   ui.NewList(),
		ifEIn:    ui.NewList(),
		ifEOut:   ui.NewList(),
	}
}

func (w *Widgets) updateHeader(c *Client) {
	var (
		h = fmt.Sprintf("──[ myLG ]── Quick NMS SNMP - %s ",
			c.SNMP.Host,
		)
		m = "Press [q] to quit"
	)

	h = h + strings.Repeat(" ", ui.TermWidth()-len(h))

	w.header.Width = ui.TermWidth()
	w.header.Height = 1
	w.header.Y = 1
	w.header.Text = h
	w.header.TextBgColor = ui.ColorCyan
	w.header.TextFgColor = ui.ColorBlack
	w.header.Border = false

	w.menu.Width = ui.TermWidth()
	w.menu.Height = 1
	w.menu.Y = 1
	w.menu.Text = m
	w.menu.TextFgColor = ui.ColorDefault
	w.menu.Border = false

	ui.Render(ui.Body)
}

func (c *Client) snmpShowInterfaceTermUI(filter string, flag map[string]interface{}) error {
	var (
		s1, s2 [][]string
		idxs   []int
		err    error
	)

	ui.DefaultEvtStream = ui.NewEvtStream()
	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()
	w := initWidgets()

	wList := []*ui.List{
		w.ifName,
		w.ifStatus,
		w.ifDescr,
		w.ifTIn,
		w.ifTOut,
		w.ifPIn,
		w.ifPOut,
		w.ifDIn,
		w.ifDOut,
		w.ifEIn,
		w.ifEOut,
	}

	for _, l := range wList {
		l.Items = make([]string, 65)
		l.X = 0
		l.Y = 0
		l.Height = 40
		l.Border = false
	}

	if len(strings.TrimSpace(filter)) > 1 {
		idxs = c.snmpGetIdx(filter)
	}

	s1, err = c.snmpGetInterfaces(idxs)
	if err != nil {
		return err
	}
	if len(s1)-1 < 1 {
		return fmt.Errorf("could not find any interface")
	}

	for i, v := range s1[0] {
		wList[i].Items[0] = fmt.Sprintf("[%s](fg-magenta,fg-bold)", v)
	}

	for i, v := range s1[1:] {
		w.ifName.Items[i+1] = v[0]
		w.ifStatus.Items[i+1] = ifStatus(v[1])
		for _, l := range wList[3:] {
			l.Items[i+1] = "-"
		}
	}

	w.updateHeader(c)

	screen := []*ui.Row{
		ui.NewRow(
			ui.NewCol(12, 0, w.header),
		),
		ui.NewRow(
			ui.NewCol(12, 0, w.menu),
		),
		ui.NewRow(
			ui.NewCol(1, 0, w.ifName),
			ui.NewCol(1, 0, w.ifStatus),
			ui.NewCol(2, 0, w.ifDescr),
			ui.NewCol(1, 0, w.ifTIn),
			ui.NewCol(1, 0, w.ifTOut),
			ui.NewCol(1, 0, w.ifPIn),
			ui.NewCol(1, 0, w.ifPOut),
			ui.NewCol(1, 0, w.ifDIn),
			ui.NewCol(1, 0, w.ifDOut),
			ui.NewCol(1, 0, w.ifEIn),
			ui.NewCol(1, 0, w.ifEOut),
		),
	}

	ui.Handle("/timer/1s", func(e ui.Event) {
		t := e.Data.(ui.EvtTimer)
		if t.Count%10 != 0 {
			return
		}

		s2, err = c.snmpGetInterfaces(idxs)
		if err != nil {
			ui.StopLoop()
		}

		for i := range s2[1:] {
			rows := normalize(s1[i+1], s2[i+1], 10)
			for c := range wList {
				wList[c].Items[i+1] = rows[c]
			}
		}

		copy(s1, s2)
		ui.Render(ui.Body)
	})

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	ui.Body.AddRows(screen...)
	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Loop()
	return nil
}