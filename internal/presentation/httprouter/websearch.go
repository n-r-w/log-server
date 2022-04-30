package httprouter

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
)

const (
	// JS для поиска
	searchJS = `
function doSearch()
{ 
	var dateFrom = document.getElementById("dateFrom").value
	var dateTo = document.getElementById("dateTo").value

	window.location.replace("/search?from=" + dateFrom.toString() + "&to=" + dateTo.toString())
}

window.addEventListener("load", function(){    
	var queryString = window.location.search;
	var urlParams = new URLSearchParams(queryString)

	var dateFrom = urlParams.get("from")
	var dateTo = urlParams.get("to")
	
	document.getElementById("dateFrom").value = dateFrom
	document.getElementById("dateTo").value = dateTo
});
`
)

var (
	columnAttrs = []g.Node{g.Attr("scope", "col"), StyleAttr("text-align: center")}
	columnClass = `px-6 py-4 font-medium text-white whitespace-nowrap`
	tableInfo   = []struct {
		headerName  string
		headerAttrs []g.Node
		headeStyle  []g.Node
		headerClass string

		columnClass string
		columnAttrs []g.Node
		columnWidth int
	}{
		{
			headerName:  "Время",
			headerAttrs: columnAttrs,
			headeStyle:  []g.Node{StyleAttr("min-width:150px; text-align: center")},
			headerClass: "px-6 py-3",
			columnClass: columnClass,
			columnAttrs: columnAttrs,
			columnWidth: 10,
		},
		{
			headerName:  "Уровень",
			headerAttrs: columnAttrs,
			headerClass: "px-6 py-3",
			columnClass: columnClass,
			columnAttrs: columnAttrs,
			columnWidth: 5,
		},
		{
			headerName:  "Информация",
			headerAttrs: columnAttrs,
			headerClass: "px-6 py-3",
			columnClass: columnClass,
			columnAttrs: columnAttrs,
			columnWidth: 80,
		},
	}

	tableClass       = Class("relative w-full text-sm text-left text-gray-400")
	tableHeaderClass = Class("text-xs uppercase bg-gray-700 text-gray-400")
	tableBodyClass   = Class("border-b bg-gray-800 border-gray-700")
	tableDivClass    = Class("relative overflow-x-auto shadow-md sm:rounded-lg")

	requestTimeFormat = "2006-01-02T15:04"
	userTimeFormat    = "02.01.2006 15:04:05"
)

func (router *HTTPRouter) webIndex(w http.ResponseWriter, r *http.Request) g.Node {
	if u, _, _ := router.isAuthenticated(r); u == nil {
		return router.renderNotLoginGeneral()
	}

	timeFromRequest := r.URL.Query().Get("from")
	timeToRequest := r.URL.Query().Get("to")

	var timeFrom time.Time
	var timeTo time.Time

	var err error

	if len(timeFromRequest) > 0 && len(timeToRequest) > 0 {
		timeFrom, err = time.Parse(requestTimeFormat, timeFromRequest)
		if err == nil {
			timeTo, err = time.Parse(requestTimeFormat, timeToRequest)
		}
	} else {
		err = errors.New("Не указан интервал дат")
	}

	var tableHeaders []g.Node
	for _, header := range tableInfo {
		tableHeaders = append(tableHeaders,
			Th(Class("sticky top-0 px-6 py-3 text-xs uppercase bg-gray-700 text-white"),
				g.Group(header.headeStyle),
				g.Attr("width", fmt.Sprintf("%d%%", header.columnWidth)),
				g.Group(header.headerAttrs), Class(header.headerClass), tableHeaderColorStyleAttr,
				g.Text(header.headerName)))
	}

	var logRecords *[]model.LogRecord
	var limited bool
	if err == nil {
		logRecords, limited, err = router.domain.LogUsecase.Find(timeFrom, timeTo, config.AppConfig.MaxLogRecordsResultWeb)
	}
	if logRecords == nil {
		logRecords = &[]model.LogRecord{}
	}

	tableRows := make([]g.Node, len(*logRecords))
	for _, record := range *logRecords {
		var rowItems []g.Node
		for hnum, header := range tableInfo {
			var cellName string
			switch hnum {
			case 0:
				cellName = record.LogTime.Format(userTimeFormat)
			case 1:
				cellName = fmt.Sprintf("%d", record.Level)
			case 2:
				cellName = record.Message1
			default:
				log.Panicln("internal error")
			}

			rowItems = append(rowItems,
				Td(
					g.Group(header.columnAttrs),
					Class(header.columnClass), tableColorStyleAttr,
					g.Text(cellName)))
		}
		tableRows = append(tableRows, Tr(Class(`border-b bg-gray-800 border-gray-700`),
			tableColorStyleAttr,
			g.Group(rowItems)))
	}

	table := Div(Class("flex flex-col "), StyleAttr("height:88vh"),
		Div(Class("flex-grow overflow-auto"),
			Table(tableClass, colorStyleAttr,
				THead(tableHeaderClass, tableHeaderColorStyleAttr,
					Tr(g.Group(tableHeaders)),
					TBody(tableBodyClass, tableColorStyleAttr, g.Group(tableRows))))))

	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	} else if limited {
		errorMessage = fmt.Sprintf("Слишком много записей, показано %d", config.AppConfig.MaxLogRecordsResultWeb)
	}

	searchParams := Div(Class("flex flex-wrap items-center space-x-0"),
		Div(textClassRowSameLine, g.Text("Время с")),
		Div(Input(buttonClassRowSameLine, ID("dateFrom"), Type("datetime-local"))),
		Div(textClassRowSameLine, g.Text("по")),
		Div(Input(buttonClassRowSameLine, ID("dateTo"), Type("datetime-local"))),
		Div(Input(buttonClassRowSameLine,
			ID("search"), Type("button"), Value("Поиск"), g.Attr("onclick", "doSearch()"))),
		Div(Label(ID("searchMessage"), Div(Class("text-red-300"), g.Text(errorMessage)))),
		Script(g.Raw(searchJS)),
	)

	return Div(searchParams, Div(tableDivClass, colorStyleAttr, table))
}
