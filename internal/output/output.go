package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

const (
	ExitOK      = 0
	ExitGeneral = 1
	ExitConfig  = 2
	ExitAPI     = 3
	ExitNetwork = 4
)

type ListResult struct {
	Data       any  `json:"data"`
	TotalCount int  `json:"total_count"`
	HasMore    bool `json:"has_more"`
}

type Printer struct {
	Format string
	Out    io.Writer
	ErrOut io.Writer
}

func NewPrinter(format string) *Printer {
	return &Printer{
		Format: format,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

func (p *Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (p *Printer) PrintList(data any, totalCount int, hasMore bool) error {
	result := ListResult{
		Data:       data,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}
	if p.Format == "table" {
		return p.printTable(data)
	}
	return p.PrintJSON(result)
}

func (p *Printer) PrintSingle(data any) error {
	if p.Format == "table" {
		return p.printTable(data)
	}
	return p.PrintJSON(data)
}

// FormatError writes the error message to ErrOut without exiting.
func (p *Printer) FormatError(msg string) {
	if p.Format == "json" {
		errObj := map[string]string{"error": msg}
		data, _ := json.Marshal(errObj)
		fmt.Fprintln(p.ErrOut, string(data))
	} else {
		fmt.Fprintf(p.ErrOut, "Error: %s\n", msg)
	}
}

func (p *Printer) PrintError(msg string, exitCode int) {
	p.FormatError(msg)
	os.Exit(exitCode)
}

type TableRenderer interface {
	TableHeaders() []string
	TableRow() []string
}

func (p *Printer) printTable(data any) error {
	w := tabwriter.NewWriter(p.Out, 0, 0, 2, ' ', 0)

	switch v := data.(type) {
	case []TableRenderer:
		if len(v) == 0 {
			fmt.Fprintln(p.Out, "No results found.")
			return nil
		}
		headers := v[0].TableHeaders()
		for i, h := range headers {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, h)
		}
		fmt.Fprintln(w)
		for _, item := range v {
			row := item.TableRow()
			for i, cell := range row {
				if i > 0 {
					fmt.Fprint(w, "\t")
				}
				fmt.Fprint(w, cell)
			}
			fmt.Fprintln(w)
		}
	case TableRenderer:
		headers := v.TableHeaders()
		for i, h := range headers {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, h)
		}
		fmt.Fprintln(w)
		row := v.TableRow()
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, cell)
		}
		fmt.Fprintln(w)
	default:
		return p.PrintJSON(data)
	}

	return w.Flush()
}

func ToTableRenderers[T TableRenderer](items []T) []TableRenderer {
	result := make([]TableRenderer, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
