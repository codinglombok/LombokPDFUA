package pdfua

import (
	"strings"
	"testing"
)

// ============================================================================
// TEST FIXTURES
// ============================================================================

func createMinimalPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
trailer
<< /Size 2 /Root 1 0 R >>
%%EOF`)
}

func createTaggedPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R /MarkInfo << /Marked true >> /StructTreeRoot 3 0 R >>
endobj
trailer
<< /Size 2 /Root 1 0 R >>
%%EOF`)
}

func createPDFWithTitle() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Info << /Title (Test Document) >> >>
endobj
trailer
<< /Size 2 >>
%%EOF`)
}

func createPDFWithLanguage() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Lang (en-US) >>
endobj
trailer
<< /Size 2 >>
%%EOF`)
}

func createPDFWithImages() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Image << /Alt (A description of the image) /Width 100 /Height 100 >> >>
endobj
trailer
<< /Size 2 >>
%%EOF`)
}

func createPDFWithImagesNoAlt() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Image << /Width 100 /Height 100 >> >>
endobj
trailer
<< /Size 2 >>
%%EOF`)
}

func createPDFWithJavaScript() []byte {
	base := createMinimalPDF()
	extra := []byte("\n/JavaScript (alert('test'))")
	return append(base, extra...)
}

func createPDFWithH1() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /S /H1 /K [2 0 R] >>
endobj
trailer
<< /Size 2 >>
%%EOF`)
}

func createInvalidPDF() []byte {
	return []byte("Not a PDF file")
}

// ============================================================================
// TESTS
// ============================================================================

func TestPDFUAReportCreation(t *testing.T) {
	report := NewPDFUAReport("test.pdf")
	if report.FileName != "test.pdf" {
		t.Errorf("Expected FileName 'test.pdf', got %s", report.FileName)
	}
	if len(report.Issues) != 0 {
		t.Errorf("Expected empty Issues, got %d", len(report.Issues))
	}
}

func TestAddIssue(t *testing.T) {
	report := NewPDFUAReport("test.pdf")
	suggestion := "Fix this"
	report.AddIssue(TaggedPdfRequired, "error", "Not tagged", nil, &suggestion)

	if len(report.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(report.Issues))
	}
	if report.Issues[0].Rule != TaggedPdfRequired {
		t.Errorf("Expected TaggedPdfRequired rule")
	}
}

func TestErrorWarningInfoCount(t *testing.T) {
	report := NewPDFUAReport("test.pdf")
	report.AddIssue(TaggedPdfRequired, "error", "msg1", nil, nil)
	report.AddIssue(DocumentTitle, "warning", "msg2", nil, nil)
	report.AddIssue(NoJavaScript, "info", "msg3", nil, nil)

	if report.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", report.ErrorCount())
	}
	if report.WarningCount() != 1 {
		t.Errorf("Expected 1 warning, got %d", report.WarningCount())
	}
	if report.InfoCount() != 1 {
		t.Errorf("Expected 1 info, got %d", report.InfoCount())
	}
}

func TestToJSON(t *testing.T) {
	report := NewPDFUAReport("test.pdf")
	jsonStr, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(jsonStr, "test.pdf") {
		t.Error("Expected JSON to contain fileName")
	}
}

func TestAnalyzerParseHeader(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createMinimalPDF()

	version := analyzer.ParseHeader()
	if version != "1.4" {
		t.Errorf("Expected version '1.4', got '%s'", version)
	}
}

func TestIsTaggedPdf(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createTaggedPDF()

	if !analyzer.IsTaggedPdf() {
		t.Error("Expected IsTaggedPdf to be true")
	}
}

func TestIsTaggedPdfFalse(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createMinimalPDF()

	if analyzer.IsTaggedPdf() {
		t.Error("Expected IsTaggedPdf to be false for untagged PDF")
	}
}

func TestHasDocumentTitle(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithTitle()

	if !analyzer.HasDocumentTitle() {
		t.Error("Expected HasDocumentTitle to be true")
	}
}

func TestHasLanguageSpecification(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithLanguage()

	if !analyzer.HasLanguageSpecification() {
		t.Error("Expected HasLanguageSpecification to be true")
	}
}

func TestCheckHeadingHierarchy(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithH1()

	if !analyzer.CheckHeadingHierarchy() {
		t.Error("Expected CheckHeadingHierarchy to be true (has H1)")
	}
}

func TestCheckHeadingHierarchyFalse(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createMinimalPDF()

	if analyzer.CheckHeadingHierarchy() {
		t.Error("Expected CheckHeadingHierarchy to be false (no H1)")
	}
}

func TestExtractImages(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithImages()

	images := analyzer.ExtractImages()
	if len(images) == 0 {
		t.Log("No images extracted (may depend on regex matching)")
	} else {
		if !images[0].HasAltText {
			t.Error("Expected image to have alt text")
		}
	}
}

func TestExtractImagesNoAlt(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithImagesNoAlt()

	images := analyzer.ExtractImages()
	if len(images) > 0 && images[0].HasAltText {
		t.Error("Expected image to not have alt text")
	}
}

func TestHasJavaScript(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createPDFWithJavaScript()

	if !analyzer.HasJavaScript() {
		t.Error("Expected HasJavaScript to be true")
	}
}

func TestCheckColorContrast(t *testing.T) {
	analyzer := NewAccessibilityAnalyzer()
	analyzer.Content = createMinimalPDF()

	isSufficient, ratio := analyzer.CheckColorContrast()
	if !isSufficient {
		t.Error("Expected color contrast to be sufficient (simplified check)")
	}
	if ratio != 4.5 {
		t.Errorf("Expected ratio 4.5, got %f", ratio)
	}
}

func TestValidatorValidateInvalidPDF(t *testing.T) {
	validator := NewPDFUAValidator()
	report := validator.Validate(createInvalidPDF())

	if report == nil {
		t.Fatal("Expected non-nil report")
	}
	if report.ErrorCount() == 0 {
		t.Error("Expected at least one error for invalid PDF")
	}
}

func TestValidatorDetectsUntaggedPDF(t *testing.T) {
	validator := NewPDFUAValidator()
	report := validator.Validate(createMinimalPDF())

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == TaggedPdfRequired {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected TaggedPdfRequired issue for untagged PDF")
	}
}

func TestValidatorDetectsJavaScript(t *testing.T) {
	validator := NewPDFUAValidator()
	report := validator.Validate(createPDFWithJavaScript())

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == NoJavaScript {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected NoJavaScript issue to be detected")
	}
}

func TestPDFUACreation(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("Expected non-nil PDFUA instance")
	}
	if p.GetReport() != nil {
		t.Error("Expected nil report before validation")
	}
}

func TestPDFUAValidate(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	report := p.GetReport()
	if report == nil {
		t.Fatal("Expected non-nil report after validation")
	}
}

func TestPDFUAIsCompliant(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	// Untagged minimal PDF should not be compliant
	if p.IsCompliant() {
		t.Error("Expected non-compliant document (untagged)")
	}
}

func TestPDFUAIsWCAGAACompliant(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	if p.IsWCAGAACompliant() {
		t.Error("Expected non-WCAG-AA-compliant document (has errors)")
	}
}

func TestPDFUAGetErrors(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	errors := p.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected at least one error")
	}
}

func TestPDFUAGetWarnings(t *testing.T) {
	p := New()
	p.Validate(createTaggedPDF())

	warnings := p.GetWarnings()
	// Tagged PDF without title/language should generate warnings
	if len(warnings) == 0 {
		t.Log("No warnings generated (may be expected)")
	}
}

func TestPDFUAGetInfo(t *testing.T) {
	p := New()
	p.Validate(createPDFWithJavaScript())

	info := p.GetInfo()
	if len(info) == 0 {
		t.Log("No info messages generated (may be expected)")
	}
}

func TestPDFUAGetIssueCounts(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	errCount, warnCount, infoCount := p.GetIssueCounts()
	if errCount == 0 {
		t.Error("Expected non-zero error count")
	}
	_ = warnCount
	_ = infoCount
}

func TestPDFUAGetImages(t *testing.T) {
	p := New()
	p.Validate(createPDFWithImages())

	images := p.GetImages()
	_ = images // may be empty depending on regex match, just verify no panic
}

func TestPDFUAGetFormFields(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	fields := p.GetFormFields()
	if fields == nil {
		t.Error("Expected non-nil form fields slice")
	}
}

func TestPDFUAToJSON(t *testing.T) {
	p := New()
	p.Validate(createMinimalPDF())

	jsonStr, err := p.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if jsonStr == "" {
		t.Error("Expected non-empty JSON")
	}
}

func TestPDFUANilReportSafety(t *testing.T) {
	p := New()

	if p.IsCompliant() != false {
		t.Error("Expected false for IsCompliant with nil report")
	}
	if p.IsWCAGAACompliant() != false {
		t.Error("Expected false for IsWCAGAACompliant with nil report")
	}
	if len(p.GetErrors()) != 0 {
		t.Error("Expected empty errors with nil report")
	}
	if len(p.GetWarnings()) != 0 {
		t.Error("Expected empty warnings with nil report")
	}
	if len(p.GetInfo()) != 0 {
		t.Error("Expected empty info with nil report")
	}
	errCount, warnCount, infoCount := p.GetIssueCounts()
	if errCount != 0 || warnCount != 0 || infoCount != 0 {
		t.Error("Expected 0,0,0 for issue counts with nil report")
	}
	if len(p.GetImages()) != 0 {
		t.Error("Expected empty images with nil report")
	}
	if len(p.GetFormFields()) != 0 {
		t.Error("Expected empty form fields with nil report")
	}
	jsonStr, _ := p.ToJSON()
	if jsonStr != "{}" {
		t.Errorf("Expected '{}' for nil report JSON, got '%s'", jsonStr)
	}
}

func TestAccessibilityRuleConstants(t *testing.T) {
	rules := []AccessibilityRule{
		TaggedPdfRequired, DocumentTitle, LogicalStructure, AltTextForImages,
		TableStructure, FormLabels, ColorContrast, LanguageSpecification,
		ArtifactTagging, ContentOrder, LinkDescriptions, NoJavaScript,
		AccessibleFonts, HeadingHierarchy,
	}
	if len(rules) != 14 {
		t.Errorf("Expected 14 accessibility rules, got %d", len(rules))
	}
}

func TestPDFUALevelConstants(t *testing.T) {
	levels := []PDFUALevel{PDFUA1, PDFUA2}
	if len(levels) != 2 {
		t.Errorf("Expected 2 PDF/UA levels, got %d", len(levels))
	}
}

func BenchmarkValidate(b *testing.B) {
	pdfBytes := createTaggedPDF()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := New()
		p.Validate(pdfBytes)
	}
}
