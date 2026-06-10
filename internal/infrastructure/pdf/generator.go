// Package pdf provides PDF generation for corporate subscription documents.
package pdf

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/DejaVuSans.ttf
var dejaVuSansFont []byte

// GroupInfo holds data for one corporate group in the PDF.
type GroupInfo struct {
	Number      int
	Name        string
	TeacherLink string
	StudentLink string
}

// PlanSummary holds purchase-level data for the PDF header.
type PlanSummary struct {
	Groups      int
	Semesters   int
	TotalAmount int
}

// GenerateFunc is the function signature for PDF generation (allows injection for testing).
type GenerateFunc func(summary PlanSummary, groups []GroupInfo) ([]byte, error)

// Generate produces the corporate subscription PDF as bytes.
func Generate(summary PlanSummary, groups []GroupInfo) ([]byte, error) {
	f := fpdf.New("P", "mm", "A4", "")
	f.AddUTF8FontFromBytes("DejaVu", "", dejaVuSansFont)

	setFont := func(size float64) {
		f.SetFont("DejaVu", "", size)
	}

	accessMonths := summary.Semesters * 5

	f.AddPage()

	// === Заголовок ===
	setFont(20)
	f.CellFormat(0, 14, "AnatBot — Корпоративный пакет", "", 1, "C", false, 0, "")
	f.Ln(2)
	f.SetDrawColor(180, 180, 180)
	f.Line(10, f.GetY(), 200, f.GetY())
	f.Ln(5)

	// === О боте ===
	setFont(13)
	f.CellFormat(0, 8, "О боте AnatBot", "", 1, "", false, 0, "")
	setFont(10)
	f.MultiCell(0, 6,
		"AnatBot — это Telegram-бот для подготовки к экзамену по анатомии. "+
			"Бот использует метод мнемоник (образных ассоциаций) для быстрого и прочного запоминания "+
			"сложных анатомических терминов и структур.\n\n"+
			"Для студентов: изучение тем с мнемоническими карточками, прохождение тестов, "+
			"отслеживание личного прогресса, последовательное открытие модулей.\n\n"+
			"Для преподавателей: мониторинг успеваемости своей группы, просмотр прогресса "+
			"каждого студента, общая статистика по группе.",
		"", "", false)
	f.Ln(5)

	// === Параметры пакета ===
	f.SetDrawColor(180, 180, 180)
	f.Line(10, f.GetY(), 200, f.GetY())
	f.Ln(5)
	setFont(13)
	f.CellFormat(0, 8, "Параметры пакета", "", 1, "", false, 0, "")
	setFont(10)
	f.CellFormat(0, 6, fmt.Sprintf("Количество групп: %d", summary.Groups), "", 1, "", false, 0, "")
	f.CellFormat(0, 6, fmt.Sprintf("Количество семестров: %d", summary.Semesters), "", 1, "", false, 0, "")
	f.CellFormat(0, 6,
		fmt.Sprintf("Срок доступа: %d × 5 мес. = %d мес. на каждую группу", summary.Semesters, accessMonths),
		"", 1, "", false, 0, "")
	f.CellFormat(0, 6, fmt.Sprintf("Стоимость: %d руб.", summary.TotalAmount), "", 1, "", false, 0, "")
	f.Ln(5)

	// === Ссылки по группам ===
	f.SetDrawColor(180, 180, 180)
	f.Line(10, f.GetY(), 200, f.GetY())
	f.Ln(5)
	setFont(13)
	f.CellFormat(0, 8, "Ссылки для активации", "", 1, "", false, 0, "")
	f.Ln(2)

	for _, g := range groups {
		setFont(11)
		f.CellFormat(0, 7, fmt.Sprintf("Группа %d: %s", g.Number, g.Name), "", 1, "", false, 0, "")
		setFont(9)
		f.CellFormat(55, 6, "Ссылка для преподавателя:", "", 0, "", false, 0, "")
		f.MultiCell(0, 6, g.TeacherLink, "", "", false)
		f.CellFormat(55, 6, "Ссылка для студентов:", "", 0, "", false, 0, "")
		f.MultiCell(0, 6, g.StudentLink, "", "", false)
		f.Ln(3)
	}

	// === Инструкция по активации ===
	f.SetDrawColor(180, 180, 180)
	f.Line(10, f.GetY(), 200, f.GetY())
	f.Ln(5)
	setFont(13)
	f.CellFormat(0, 8, "Инструкция по активации", "", 1, "", false, 0, "")
	setFont(10)
	f.MultiCell(0, 6,
		"Для преподавателей:\n"+
			"1. Найдите в этом документе персональную ссылку для преподавателя вашей группы.\n"+
			"2. Откройте Telegram и перейдите по ссылке — запустится бот AnatBot.\n"+
			"3. Нажмите «Начать». Ваш аккаунт будет автоматически привязан к группе.\n"+
			fmt.Sprintf("4. Вы получите доступ к учебному контенту на %d мес. и инструменты просмотра прогресса группы.\n\n",
				accessMonths)+
			"Для студентов:\n"+
			"1. Получите от преподавателя ссылку для студентов вашей группы.\n"+
			"2. Перейдите по ссылке и нажмите «Начать» в боте Mnemo.\n"+
			fmt.Sprintf("3. Ваша подписка на %d мес. активируется автоматически (до 30 студентов в группе).",
				accessMonths),
		"", "", false)

	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}
