// Package pdfua implements LombokPDFUA - Go Port (v0.9.0-alpha)
// PDF/UA accessibility validation (ISO 14289 standard).
// Supports WCAG 2.1 AA compliance checking.
// Apache 2.0 License | Multi-language support
package pdfua

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"
)

// ============================================================================
// PDF/UA CONSTANTS
// ============================================================================

// PDFUALevel represents PDF/UA compliance levels
type PDFUALevel string

const (
	PDFUA1 PDFUALevel = "PDF/UA-1" // 2012, ISO 14289-1
	PDFUA2 PDFUALevel = "PDF/UA-2" // 2020, ISO 14289-2
)

// AccessibilityRule represents PDF/UA accessibility rules
type AccessibilityRule string

const (
	TaggedPdfRequired       AccessibilityRule = "tagged_pdf_required"
	DocumentTitle           AccessibilityRule = "document_title"
	LogicalStructure        AccessibilityRule = "logical_structure"
	AltTextForImages        AccessibilityRule = "alt_text_for_images"
	TableStructure          AccessibilityRule = "table_structure"
	FormLabels              AccessibilityRule = "form_labels"
	ColorContrast           AccessibilityRule = "color_contrast"
	LanguageSpecification   AccessibilityRule = "language_specification"
	ArtifactTagging         AccessibilityRule = "artifact_tagging"
	ContentOrder            AccessibilityRule = "content_order"
	LinkDescriptions        AccessibilityRule = "link_descriptions"
	NoJavaScript            AccessibilityRule = "no_javascript"
	AccessibleFonts         AccessibilityRule = "accessible_fonts"
	HeadingHierarchy        AccessibilityRule = "heading_hierarchy"
)

// ============================================================================
// PDF/UA DATA STRUCTURES
// ============================================================================

// AccessibilityIssue represents an accessibility issue
type AccessibilityIssue struct {
	Rule       AccessibilityRule `json:"rule"`
	Severity   string            `json:"severity"` // "error", "warning", "info"
	Message    string            `json:"message"`
	Location   *string           `json:"location,omitempty"`
	Suggestion *string           `json:"suggestion,omitempty"`
}

// StructureNode represents a document structure tree node
type StructureNode struct {
	Tag      string          `json:"tag"`
	Role     *string         `json:"role,omitempty"`
	Title    *string         `json:"title,omitempty"`
	AltText  *string         `json:"altText,omitempty"`
	Children []StructureNode `json:"children"`
}

// FormField represents form field information
type FormField struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"` // Text, Button, Checkbox, Radio, etc.
	HasLabel   bool    `json:"hasLabel"`
	IsRequired bool    `json:"isRequired"`
	Tooltip    *string `json:"tooltip,omitempty"`
}

// ImageInfo represents image information
type ImageInfo struct {
	ID           string  `json:"id"`
	HasAltText   bool    `json:"hasAltText"`
	AltText      *string `json:"altText,omitempty"`
	IsDecorative bool    `json:"isDecorative"`
}

// PDFUAReport represents a PDF/UA compliance report
type PDFUAReport struct {
	FileName          string                `json:"fileName"`
	DetectedLevel     *PDFUALevel           `json:"detectedLevel,omitempty"`
	IsCompliant       bool                  `json:"isCompliant"`
	Issues            []AccessibilityIssue  `json:"issues"`
	StructureTree     *StructureNode        `json:"structureTree,omitempty"`
	Images            []ImageInfo           `json:"images"`
	FormFields        []FormField           `json:"formFields"`
	PageCount         int                   `json:"pageCount"`
	FileSize          int                   `json:"fileSize"`
	AnalysisTimestamp string                `json:"analysisTimestamp"`
	WCAGAACompliant   bool                  `json:"wcagAACompliant"`
}

// NewPDFUAReport creates a new PDFUAReport with defaults
func NewPDFUAReport(fileName string) *PDFUAReport {
	return &PDFUAReport{
		FileName:   fileName,
		Issues:     []AccessibilityIssue{},
		Images:     []ImageInfo{},
		FormFields: []FormField{},
	}
}

// AddIssue adds an accessibility issue to the report
func (r *PDFUAReport) AddIssue(rule AccessibilityRule, severity, message string, location, suggestion *string) {
	r.Issues = append(r.Issues, AccessibilityIssue{
		Rule:       rule,
		Severity:   severity,
		Message:    message,
		Location:   location,
		Suggestion: suggestion,
	})
}

// ErrorCount returns the number of error-severity issues
func (r *PDFUAReport) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warning-severity issues
func (r *PDFUAReport) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "warning" {
			count++
		}
	}
	return count
}

// InfoCount returns the number of info-severity issues
func (r *PDFUAReport) InfoCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "info" {
			count++
		}
	}
	return count
}

// ToJSON converts the report to a JSON string
func (r *PDFUAReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ============================================================================
// PDF ACCESSIBILITY ANALYZER
// ============================================================================

// AccessibilityAnalyzer analyzes PDF for accessibility
type AccessibilityAnalyzer struct {
	Content []byte
}

// NewAccessibilityAnalyzer creates a new AccessibilityAnalyzer
func NewAccessibilityAnalyzer() *AccessibilityAnalyzer {
	return &AccessibilityAnalyzer{}
}

var pdfuaVersionRe = regexp.MustCompile(`%PDF-(\d\.\d)`)

// ParseHeader parses PDF version
func (a *AccessibilityAnalyzer) ParseHeader() string {
	if a.Content == nil {
		return ""
	}

	end := 1024
	if len(a.Content) < end {
		end = len(a.Content)
	}
	header := string(a.Content[:end])

	matches := pdfuaVersionRe.FindStringSubmatch(header)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// IsTaggedPdf checks if PDF is tagged
func (a *AccessibilityAnalyzer) IsTaggedPdf() bool {
	if a.Content == nil {
		return false
	}
	content := string(a.Content)
	return contains(content, "/StructTreeRoot") || contains(content, "/StructParent")
}

var titleCheckRe = regexp.MustCompile(`/Title\s*\(`)

// HasDocumentTitle checks if document has title
func (a *AccessibilityAnalyzer) HasDocumentTitle() bool {
	if a.Content == nil {
		return false
	}
	return titleCheckRe.MatchString(string(a.Content))
}

// HasLanguageSpecification checks if language is specified
func (a *AccessibilityAnalyzer) HasLanguageSpecification() bool {
	if a.Content == nil {
		return false
	}
	return contains(string(a.Content), "/Lang")
}

var h1Re = regexp.MustCompile(`/H1\b`)

// CheckHeadingHierarchy checks if heading hierarchy is correct
func (a *AccessibilityAnalyzer) CheckHeadingHierarchy() bool {
	if a.Content == nil {
		return false
	}

	content := string(a.Content)
	h1Matches := h1Re.FindAllString(content, -1)

	// Valid hierarchy: at least one H1 present
	return len(h1Matches) > 0
}

var imgBlockRe = regexp.MustCompile(`(?s)/Image.*?<<(.*?)>>`)
var altTextRe = regexp.MustCompile(`/Alt\s*\((.*?)\)`)

// ExtractImages extracts image information
func (a *AccessibilityAnalyzer) ExtractImages() []ImageInfo {
	images := []ImageInfo{}
	if a.Content == nil {
		return images
	}

	content := string(a.Content)
	matches := imgBlockRe.FindAllStringSubmatch(content, -1)

	for idx, match := range matches {
		imgBlock := match[1]

		var altText *string
		hasAlt := false
		if altMatch := altTextRe.FindStringSubmatch(imgBlock); len(altMatch) > 1 {
			altText = &altMatch[1]
			hasAlt = true
		}

		isDecorative := contains(imgBlock, "/Artifact")

		images = append(images, ImageInfo{
			ID:           fmt.Sprintf("Image-%d", idx+1),
			HasAltText:   hasAlt,
			AltText:      altText,
			IsDecorative: isDecorative,
		})
	}

	return images
}

var formFieldRe = regexp.MustCompile(`(?s)/FT\s*/(\w+).*?/T\s*\((.*?)\)`)

// ExtractFormFields extracts form field information
func (a *AccessibilityAnalyzer) ExtractFormFields() []FormField {
	fields := []FormField{}
	if a.Content == nil {
		return fields
	}

	content := string(a.Content)
	matches := formFieldRe.FindAllStringSubmatch(content, -1)

	hasLabel := contains(content, "/Lbl")
	isRequired := contains(content, "/Ff")

	for _, match := range matches {
		fieldType := match[1]
		fieldName := match[2]

		fields = append(fields, FormField{
			Name:       fieldName,
			Type:       fieldType,
			HasLabel:   hasLabel,
			IsRequired: isRequired,
		})
	}

	return fields
}

// HasJavaScript checks for JavaScript
func (a *AccessibilityAnalyzer) HasJavaScript() bool {
	if a.Content == nil {
		return false
	}
	return contains(string(a.Content), "/JavaScript")
}

// CheckColorContrast checks color contrast (simplified check)
// Returns (isSufficient, contrastRatio)
func (a *AccessibilityAnalyzer) CheckColorContrast() (bool, float64) {
	// Simplified implementation - full implementation would analyze actual colors
	if a.Content == nil {
		return true, 4.5
	}
	return true, 4.5
}

// contains is a simple substring helper
func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}

// ============================================================================
// PDF/UA VALIDATOR
// ============================================================================

// PDFUAValidator validates PDF/UA accessibility
type PDFUAValidator struct {
	Analyzer *AccessibilityAnalyzer
}

// NewPDFUAValidator creates a new PDFUAValidator
func NewPDFUAValidator() *PDFUAValidator {
	return &PDFUAValidator{
		Analyzer: NewAccessibilityAnalyzer(),
	}
}

// Validate validates PDF/UA compliance for the given bytes
func (v *PDFUAValidator) Validate(pdfBytes []byte) *PDFUAReport {
	report := NewPDFUAReport("document.pdf")
	report.FileSize = len(pdfBytes)
	report.AnalysisTimestamp = time.Now().Format(time.RFC3339)

	v.Analyzer.Content = pdfBytes

	pdfVersion := v.Analyzer.ParseHeader()
	if pdfVersion == "" {
		suggestion := "Ensure file is a valid PDF document"
		report.AddIssue(TaggedPdfRequired, "error", "Not a valid PDF file", nil, &suggestion)
		return report
	}

	report.Images = v.Analyzer.ExtractImages()
	report.FormFields = v.Analyzer.ExtractFormFields()

	v.checkTaggedPdf(report)
	v.checkDocumentTitle(report)
	v.checkLanguageSpecification(report)
	v.checkImageAltText(report)
	v.checkFormLabels(report)
	v.checkHeadingHierarchy(report)
	v.checkNoJavaScript(report)
	v.checkColorContrast(report)

	errorCount := report.ErrorCount()
	report.IsCompliant = errorCount == 0
	report.WCAGAACompliant = errorCount == 0 && report.WarningCount() <= 2

	return report
}

func (v *PDFUAValidator) checkTaggedPdf(report *PDFUAReport) {
	if !v.Analyzer.IsTaggedPdf() {
		suggestion := "Create a tagged PDF with proper document structure"
		report.AddIssue(TaggedPdfRequired, "error", "PDF is not tagged (no structure tree)", nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkDocumentTitle(report *PDFUAReport) {
	if !v.Analyzer.HasDocumentTitle() {
		suggestion := "Add title to document properties"
		report.AddIssue(DocumentTitle, "warning", "Document title not found", nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkLanguageSpecification(report *PDFUAReport) {
	if !v.Analyzer.HasLanguageSpecification() {
		suggestion := "Set document language in PDF properties"
		report.AddIssue(LanguageSpecification, "warning", "Document language not specified", nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkImageAltText(report *PDFUAReport) {
	count := 0
	for _, img := range report.Images {
		if !img.HasAltText && !img.IsDecorative {
			count++
		}
	}
	if count > 0 {
		suggestion := "Add descriptive alt text to all non-decorative images"
		msg := fmt.Sprintf("Found %d image(s) without alt text", count)
		report.AddIssue(AltTextForImages, "error", msg, nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkFormLabels(report *PDFUAReport) {
	count := 0
	for _, f := range report.FormFields {
		if !f.HasLabel {
			count++
		}
	}
	if count > 0 {
		suggestion := "Associate labels with all form fields"
		msg := fmt.Sprintf("Found %d form field(s) without labels", count)
		report.AddIssue(FormLabels, "error", msg, nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkHeadingHierarchy(report *PDFUAReport) {
	if !v.Analyzer.CheckHeadingHierarchy() {
		suggestion := "Use proper heading hierarchy (H1 -> H2 -> H3)"
		report.AddIssue(HeadingHierarchy, "warning", "Heading hierarchy may not be correct", nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkNoJavaScript(report *PDFUAReport) {
	if v.Analyzer.HasJavaScript() {
		suggestion := "Consider removing JavaScript for better accessibility"
		report.AddIssue(NoJavaScript, "info", "PDF contains JavaScript", nil, &suggestion)
	}
}

func (v *PDFUAValidator) checkColorContrast(report *PDFUAReport) {
	isSufficient, contrastRatio := v.Analyzer.CheckColorContrast()
	if !isSufficient {
		suggestion := "Ensure text has at least 4.5:1 color contrast (WCAG AA)"
		msg := fmt.Sprintf("Color contrast may be insufficient (ratio: %.1f:1)", contrastRatio)
		report.AddIssue(ColorContrast, "warning", msg, nil, &suggestion)
	}
}

// ============================================================================
// MAIN CLASS
// ============================================================================

// PDFUA is the main PDF/UA validator
type PDFUA struct {
	validator *PDFUAValidator
	report    *PDFUAReport
}

// New creates a new PDFUA instance
func New() *PDFUA {
	return &PDFUA{
		validator: NewPDFUAValidator(),
	}
}

// Validate validates PDF/UA compliance from bytes
func (p *PDFUA) Validate(pdfBytes []byte) *PDFUA {
	p.report = p.validator.Validate(pdfBytes)
	return p
}

// ValidateFile validates PDF/UA compliance from a file path
func (p *PDFUA) ValidateFile(path string) (*PDFUA, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	p.report = p.validator.Validate(data)
	p.report.FileName = path
	return p, nil
}

// GetReport returns the validation report
func (p *PDFUA) GetReport() *PDFUAReport {
	return p.report
}

// IsCompliant checks if the document is compliant
func (p *PDFUA) IsCompliant() bool {
	if p.report == nil {
		return false
	}
	return p.report.IsCompliant
}

// IsWCAGAACompliant checks WCAG 2.1 AA compliance
func (p *PDFUA) IsWCAGAACompliant() bool {
	if p.report == nil {
		return false
	}
	return p.report.WCAGAACompliant
}

// GetErrors returns all error-severity issues
func (p *PDFUA) GetErrors() []AccessibilityIssue {
	if p.report == nil {
		return []AccessibilityIssue{}
	}
	var errors []AccessibilityIssue
	for _, issue := range p.report.Issues {
		if issue.Severity == "error" {
			errors = append(errors, issue)
		}
	}
	return errors
}

// GetWarnings returns all warning-severity issues
func (p *PDFUA) GetWarnings() []AccessibilityIssue {
	if p.report == nil {
		return []AccessibilityIssue{}
	}
	var warnings []AccessibilityIssue
	for _, issue := range p.report.Issues {
		if issue.Severity == "warning" {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// GetInfo returns all info-severity issues
func (p *PDFUA) GetInfo() []AccessibilityIssue {
	if p.report == nil {
		return []AccessibilityIssue{}
	}
	var info []AccessibilityIssue
	for _, issue := range p.report.Issues {
		if issue.Severity == "info" {
			info = append(info, issue)
		}
	}
	return info
}

// GetIssueCounts returns error, warning, and info counts
func (p *PDFUA) GetIssueCounts() (int, int, int) {
	if p.report == nil {
		return 0, 0, 0
	}
	return p.report.ErrorCount(), p.report.WarningCount(), p.report.InfoCount()
}

// GetImages returns image information
func (p *PDFUA) GetImages() []ImageInfo {
	if p.report == nil {
		return []ImageInfo{}
	}
	return p.report.Images
}

// GetFormFields returns form field information
func (p *PDFUA) GetFormFields() []FormField {
	if p.report == nil {
		return []FormField{}
	}
	return p.report.FormFields
}

// ToJSON exports the report to JSON
func (p *PDFUA) ToJSON() (string, error) {
	if p.report == nil {
		return "{}", nil
	}
	return p.report.ToJSON()
}
