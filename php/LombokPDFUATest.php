<?php
use Lombok\PDFUA;
use PHPUnit\Framework\TestCase;

class LombokPDFUATest extends TestCase {

    public function testInvalidPdfReturnsError() {
        $pdfua = new PDFUA();
        $pdfua->validate("not a pdf at all");
        $this->assertFalse($pdfua->isCompliant());
        $errors = $pdfua->getErrors();
        $this->assertCount(1, $errors);
        $this->assertEquals('tagged_pdf_required', $errors[0]->rule);
    }

    public function testUntaggedPdfFlagsError() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n%%EOF");
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertTrue(in_array('tagged_pdf_required', $rules));
    }

    public function testTaggedPdfDoesNotFlagStructure() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF");
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertFalse(in_array('tagged_pdf_required', $rules));
    }

    public function testMissingTitleFlagsWarning() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF");
        $warnings = $pdfua->getWarnings();
        $rules = array_map(fn($w) => $w->rule, $warnings);
        $this->assertTrue(in_array('document_title', $rules));
    }

    public function testPresentTitleDoesNotFlag() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n%%EOF");
        $warnings = $pdfua->getWarnings();
        $rules = array_map(fn($w) => $w->rule, $warnings);
        $this->assertFalse(in_array('document_title', $rules));
    }

    public function testMissingLanguageFlagsWarning() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n/StructTreeRoot 5 0 R\n%%EOF");
        $warnings = $pdfua->getWarnings();
        $rules = array_map(fn($w) => $w->rule, $warnings);
        $this->assertTrue(in_array('language_specification', $rules));
    }

    public function testImageWithoutAltTextFlagsError() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Width 100 /Height 100>>\n%%EOF";
        $pdfua->validate($content);
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertTrue(in_array('alt_text_for_images', $rules));
    }

    public function testImageWithAltTextDoesNotFlag() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Alt (A description)>>\n%%EOF";
        $pdfua->validate($content);
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertFalse(in_array('alt_text_for_images', $rules));

        $images = $pdfua->getImages();
        $this->assertCount(1, $images);
        $this->assertEquals('A description', $images[0]->altText);
        $this->assertTrue($images[0]->hasAltText);
    }

    public function testDecorativeImageWithoutAltDoesNotFlag() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/Image <</Artifact true>>\n%%EOF";
        $pdfua->validate($content);
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertFalse(in_array('alt_text_for_images', $rules));
    }

    public function testFormFieldWithoutLabelFlagsError() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/FT /Tx /T (username)\n%%EOF";
        $pdfua->validate($content);
        $errors = $pdfua->getErrors();
        $rules = array_map(fn($e) => $e->rule, $errors);
        $this->assertTrue(in_array('form_labels', $rules));

        $fields = $pdfua->getFormFields();
        $this->assertCount(1, $fields);
        $this->assertEquals('username', $fields[0]->name);
        $this->assertEquals('Tx', $fields[0]->type);
    }

    public function testJavaScriptFlagsInfo() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/JavaScript (app.alert(1))\n%%EOF";
        $pdfua->validate($content);
        $info = $pdfua->getInfo();
        $rules = array_map(fn($i) => $i->rule, $info);
        $this->assertTrue(in_array('no_javascript', $rules));
    }

    public function testFullyCompliantDocument() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n/Lang (en-US)\n/H1\n%%EOF";
        $pdfua->validate($content);
        $this->assertTrue($pdfua->isCompliant());
        $this->assertTrue($pdfua->isWCAGAACompliant());
        $this->assertCount(0, $pdfua->getErrors());
    }

    public function testGetIssueCounts() {
        $pdfua = new PDFUA();
        $content = "%PDF-1.7\n%%EOF"; // untagged, no title, no lang
        $pdfua->validate($content);
        [$errors, $warnings, $info] = $pdfua->getIssueCounts();
        $this->assertEquals(1, $errors); // untagged
        $this->assertTrue($warnings >= 2); // title + lang + heading
    }

    public function testFluentInterfaceReturnsSelf() {
        $pdfua = new PDFUA();
        $result = $pdfua->validate("%PDF-1.7\n%%EOF");
        $this->assertInstanceOf(PDFUA::class, $result);
        $this->assertNotNull($result->getReport());
    }

    public function testFileNameIsHardcoded() {
        // Known limitation: PDFUAReport always hardcodes "document.pdf"
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n%%EOF");
        $this->assertEquals('document.pdf', $pdfua->getReport()->fileName);
    }

    public function testToJSON() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n/Lang (en-US)\n/H1\n%%EOF");
        $json = $pdfua->toJSON();
        $decoded = json_decode($json, true);
        $this->assertTrue($decoded['isCompliant']);
    }

    public function testToDict() {
        $pdfua = new PDFUA();
        $pdfua->validate("%PDF-1.7\n%%EOF");
        $dict = $pdfua->toDict();
        $this->assertEquals('document.pdf', $dict['fileName']);
        $this->assertArrayHasKey('errorCount', $dict);
    }

    public function testBeforeValidateReturnsDefaults() {
        $pdfua = new PDFUA();
        $this->assertFalse($pdfua->isCompliant());
        $this->assertFalse($pdfua->isWCAGAACompliant());
        $this->assertCount(0, $pdfua->getErrors());
        $this->assertEquals('{}', $pdfua->toJSON());
    }

    public function testValidateFromFilePath() {
        $tmpFile = tempnam(sys_get_temp_dir(), 'lombokpdfua_') . '.pdf';
        file_put_contents($tmpFile, "%PDF-1.7\n/StructTreeRoot 5 0 R\n/Title (My Doc)\n/Lang (en-US)\n/H1\n%%EOF");
        try {
            $pdfua = new PDFUA();
            $pdfua->validate($tmpFile);
            $this->assertTrue($pdfua->isCompliant());
        } finally {
            unlink($tmpFile);
        }
    }
}
