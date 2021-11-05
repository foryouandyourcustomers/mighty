package export

import (
	"fmt"
	"github.com/elliotchance/orderedmap"
	"github.com/leanovate/mite-go/domain"
	log "github.com/sirupsen/logrus"
	"github.com/xhit/go-str2duration/v2"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const sheetSummaryName = "Summary"

type XlFile struct {
	fileName string
	file     *excelize.File
}

func ExcelFile(fileName string) *XlFile {
	return &XlFile{
		fileName,
		excelize.NewFile(),
	}
}

func (xlx *XlFile) LoadAllEntries(entries []*domain.TimeEntry) {
	log.Infof("Loading %d entries to %s", len(entries), xlx.fileName)

	textStyle, err := xlx.file.NewStyle(`{"alignment":{"horizontal": "left","vertical": "center","wrap_text": true}}`)
	if err != nil {
		fmt.Println(err)
	}

	var monthEntriesCounts map[string]int = make(map[string]int)
	monthEntriesTotalHours := orderedmap.NewOrderedMap()

	for _, entry := range entries {

		log.Debugf("Loading entry %s", entry.Id)

		entryMonth := fmt.Sprintf("%s %d", entry.Date.Month(), entry.Date.Year())

		xlx.file.NewSheet(entryMonth)
		currentRow := monthEntriesCounts[entryMonth]
		currentRow++

		if currentRow == 1 {
			xlx.WriteHeader(entryMonth, currentRow, []string{"Date", "Project Name", "Service Name", "Billable?", "Time", "Entry Description"})
			currentRow += 2
		}

		monthEntriesCounts[entryMonth] = currentRow
		min := monthEntriesTotalHours.GetOrDefault(entryMonth, 0).(int) + entry.Minutes.Value()
		monthEntriesTotalHours.Set(entryMonth, min)

		xlx.WriteEntry(entry.Id.String(), entryMonth, currentRow, []string{
			entry.Date.String(),
			fmt.Sprintf("%s", entry.ProjectName),
			fmt.Sprintf("%s", entry.ServiceName),
			strconv.FormatBool(entry.Billable),
			entry.Minutes.String(), entry.Note,
		})

		// fit cell row height
		count := strings.Count(entry.Note, "\n")
		if count > 1 {
			rowHeight, err := xlx.file.GetRowHeight(entryMonth, currentRow)

			if err != nil {
				log.Fatal(err)
			}

			rowHeight = rowHeight * float64(count)
			err = xlx.file.SetRowHeight(entryMonth, currentRow, rowHeight)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// fit the cell width
	log.Debug("Auto adjusting cell widths")

	for month, _ := range monthEntriesCounts {
		cols, err := xlx.file.GetCols(month)

		if err != nil {
			log.Fatal(err)
		}

		for colIx, col := range cols {
			maxColWidth := 0

			for _, rowCell := range col {
				cellWidth := utf8.RuneCountInString(rowCell)

				if cellWidth > maxColWidth {
					maxColWidth = cellWidth
				}
			}

			colName, err := excelize.ColumnNumberToName(colIx + 1)

			if err != nil {
				log.Fatal(err)
			}

			if maxColWidth > excelize.MaxColumnWidth {
				maxColWidth = 150
			} else {
				maxColWidth += 5
			}

			err = xlx.file.SetColWidth(month, colName, colName, float64(maxColWidth))

			if err != nil {
				log.Fatal(err)
			}
			err = xlx.file.SetColStyle(month, colName, textStyle)

			if err != nil {
				log.Fatal(err)
			}
		}

	}
	log.Debug("Writing the summary...")
	xlx.writeSummary(monthEntriesTotalHours)
}

func (xlx *XlFile) WriteHeader(sheetName string, row int, columnData []string) {
	startColumn := 'A'
	for _, d := range columnData {
		xlx.writeRichCellData(sheetName, fmt.Sprintf("%c%d", startColumn, row), []excelize.RichTextRun{
			{
				Text: d,
				Font: &excelize.Font{
					Bold: true,
				},
			}})
		startColumn++
	}
}

func (xlx *XlFile) WriteEntry(entryId, sheetName string, row int, columnData []string) {
	startColumn := 'A'
	for _, d := range columnData {
		xlx.writeCellData(sheetName, fmt.Sprintf("%c%d", startColumn, row), d)
		startColumn++
	}
	// hide this
	entryIdAxis := fmt.Sprintf("%c%d", startColumn, row)
	xlx.writeCellData(sheetName, entryIdAxis, entryId)
	err := xlx.file.SetColVisible(sheetName, string(startColumn), false)
	if err != nil {
		log.Fatal(err)
	}
}

func (xlx *XlFile) writeCellData(sheetName, axis string, cellData string) {

	if xlx.file.GetSheetIndex(sheetName) < 0 {
		xlx.file.NewSheet(sheetName)
	}

	err := xlx.file.SetCellValue(sheetName, axis, cellData)
	if err != nil {
		log.Fatal(err)
	}

}

func (xlx *XlFile) writeRichCellData(sheetName, axis string, cellData []excelize.RichTextRun) {

	if xlx.file.GetSheetIndex(sheetName) < 0 {
		xlx.file.NewSheet(sheetName)
	}

	err := xlx.file.SetCellRichText(sheetName, axis, cellData)
	if err != nil {
		log.Fatal(err)
	}
}

func (xlx *XlFile) writeSummary(totalHours *orderedmap.OrderedMap) {
	xlx.WriteHeader(sheetSummaryName, 1, []string{"Month", "Total Hours"})
	row := 3
	for _, month := range totalHours.Keys() {
		axisMonth := fmt.Sprintf("A%d", row)
		axisHours := fmt.Sprintf("B%d", row)

		xlx.writeCellData(sheetSummaryName, axisMonth, month.(string))
		totalMins := totalHours.GetOrDefault(month, 0).(int)

		err := xlx.file.SetCellHyperLink(sheetSummaryName, axisMonth, fmt.Sprintf("'%s'!%s", month, "A1"), "Location")
		if err != nil {
			log.Fatal(err)
		}

		xlx.writeCellData(sheetSummaryName, axisHours, domain.NewMinutes(totalMins).String())
		row++
	}

}

// ReloadFromDisk This is a destructive action. If you are currently working on sheet data
// this will be lost - following will re-read the file and the current changes will be lost
func (xlx *XlFile) ReloadFromDisk() error {
	log.Debug("Reloading from disk...")

	file, err := excelize.OpenFile(xlx.fileName)
	if err != nil {
		return err
	}

	xlx.file = file
	return nil
}

func (xlx *XlFile) SaveToDisk() error {
	log.Debug("Writing to disk ...")

	xlx.file.SetActiveSheet(xlx.file.GetSheetIndex(sheetSummaryName))
	return xlx.file.SaveAs(xlx.fileName)
}

func (xlx *XlFile) ReadAllEntriesBySheet(sheetName string) []domain.TimeEntry {
	log.Debugf("Reading all entries from %s sheet", sheetName)

	pmap := xlx.readProjectId()
	smap := xlx.readServiceId()

	rows, err := xlx.file.GetRows(sheetName)
	if err != nil {
		log.Fatal(err)
	}
	var timeEntries []domain.TimeEntry
	for rIx, row := range rows {
		// skip header
		if rIx > 1 {
			var entryDate domain.LocalDate
			var entryTime domain.Minutes
			var serviceId domain.ServiceId
			var projectId domain.ProjectId
			var isEntryBillable bool
			var entryNotes string
			var entryId domain.TimeEntryId

			for cIx, cellData := range row {
				cellNr := cIx + 1
				colName, err := excelize.ColumnNumberToName(cellNr)

				if err != nil {
					log.Fatal(err)
				}

				switch colName {
				case "A":
					entryDate, err = domain.ParseLocalDate(cellData)
					if err != nil {
						log.Fatal(err)
					}
				case "B":

					id, ok := pmap[strings.ToLower(cellData)]
					if !ok {
						log.Errorf("Unable to look id for project %s ", cellData)
					}
					projectId = id
				case "C":

					id, ok := smap[strings.ToLower(cellData)]

					if !ok {
						log.Errorf("Unable to look id for service %s ", cellData)
					}
					serviceId = id

				case "D":
					isEntryBillable, err = strconv.ParseBool(cellData)
					if err != nil {
						log.Fatal(err)
					}
				case "E":
					duration, err := str2duration.ParseDuration(cellData)
					if err != nil {
						log.Fatal(err)
					}
					entryTime = domain.NewMinutes(int(duration.Minutes()))
				case "F":
					entryNotes = cellData
				case "G":
					entryId, err = domain.ParseTimeEntryId(cellData)
					if err != nil {
						log.Fatal(err)
					}
				}

			}

			timeEntries = append(timeEntries, domain.TimeEntry{
				Id:        entryId,
				Minutes:   entryTime,
				Date:      entryDate,
				Note:      entryNotes,
				Billable:  isEntryBillable,
				UserId:    domain.CurrentUser,
				ProjectId: projectId,
				ServiceId: serviceId,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}

	}
	return timeEntries
}

func (xlx *XlFile) saveServiceId(serviceIdMap map[string]domain.ServiceId) error {

	log.Debug("Writing ServiceIds...")

	sheetName := "ServiceIds"
	xlx.file.NewSheet(sheetName)
	err := xlx.file.SetSheetVisible(sheetName, false)

	if err != nil {
		return err
	}
	row := 1
	xlx.WriteHeader(sheetName, row, []string{"serviceId", "serviceName"})
	row++
	for name, id := range serviceIdMap {
		xlx.writeCellData(sheetName, fmt.Sprintf("A%d", row), name)
		xlx.writeCellData(sheetName, fmt.Sprintf("B%d", row), id.String())
		row++
	}

	return nil
}

func (xlx *XlFile) readServiceId() map[string]domain.ServiceId {
	log.Debug("Reading ServiceIds...")
	sheetName := "ServiceIds"
	serviceIdMap := make(map[string]domain.ServiceId)

	rows, err := xlx.file.GetRows(sheetName)
	if err != nil {
		log.Fatal(err)
	}
	for rIx, row := range rows {
		// skip header
		if rIx > 0 {
			var serviceName string
			var serviceId domain.ServiceId
			for cIx, cellData := range row {

				cellNr := cIx + 1
				colName, err := excelize.ColumnNumberToName(cellNr)

				if err != nil {
					log.Fatal(err)
				}

				switch colName {
				case "A":
					serviceName = cellData
				case "B":

					id, err := strconv.Atoi(cellData)
					if err != nil {
						log.Fatal(err)
					}
					serviceId = domain.NewServiceId(id)

				}

			}
			serviceIdMap[strings.ToLower(serviceName)] = serviceId
			log.Debugf("found %s=%s", serviceName, serviceId)

		}

	}
	return serviceIdMap
}

func (xlx *XlFile) saveProjectId(projectIdMap map[string]domain.ProjectId) error {
	log.Debug("Writing ProjecId...")

	sheetName := "ProjectIds"
	xlx.file.NewSheet(sheetName)
	err := xlx.file.SetSheetVisible(sheetName, false)

	if err != nil {
		return err
	}
	row := 1
	xlx.WriteHeader(sheetName, row, []string{"projectId", "projectName"})
	row++
	for name, id := range projectIdMap {
		xlx.writeCellData(sheetName, fmt.Sprintf("A%d", row), name)
		xlx.writeCellData(sheetName, fmt.Sprintf("B%d", row), id.String())
		row++
	}

	return nil
}
func (xlx *XlFile) readProjectId() map[string]domain.ProjectId {
	log.Debug("Reading ProjectIds...")
	sheetName := "ProjectIds"
	projectIdMap := make(map[string]domain.ProjectId)

	rows, err := xlx.file.GetRows(sheetName)
	if err != nil {
		log.Fatal(err)
	}
	for rIx, row := range rows {
		// skip header
		if rIx > 0 {
			var projectName string
			var projectId domain.ProjectId

			for cIx, cellData := range row {

				cellNr := cIx + 1
				colName, err := excelize.ColumnNumberToName(cellNr)

				if err != nil {
					log.Fatal(err)
				}

				switch colName {
				case "A":
					projectName = cellData
				case "B":

					id, err := strconv.Atoi(cellData)
					if err != nil {
						log.Fatal(err)
					}
					projectId = domain.NewProjectId(id)

				}

			}
			projectIdMap[strings.ToLower(projectName)] = projectId
			log.Debugf("found %s=%s", projectName, projectId)
		}

	}
	return projectIdMap
}

func (xlx *XlFile) ReadAllEntries(date domain.LocalDate) []domain.TimeEntry {
	return xlx.ReadAllEntriesBySheet(fmt.Sprintf("%s %d", date.Month(), date.Year()))
}

func (xlx *XlFile) GetSheets() {
	for index, name := range xlx.file.GetSheetMap() {
		fmt.Println(index, name)
	}
}

func (xlx *XlFile) SaveAllEntries(entries []*domain.TimeEntry) error {
	xlx.LoadAllEntries(entries)
	return xlx.SaveToDisk()
}

func (xlx *XlFile) SaveServiceProjects(sMap map[string]domain.ServiceId, pMap map[string]domain.ProjectId) {
	err := xlx.saveServiceId(sMap)
	if err != nil {
		log.Fatal(err)

	}
	err = xlx.saveProjectId(pMap)
	if err != nil {
		log.Fatal(err)
	}

	err = xlx.SaveToDisk()
	if err != nil {
		log.Fatal(err)
	}

}
