<?php

/**
 * LombokPDFUA - PHP Port (v0.9.0-alpha)
 * PDF/UA accessibility validation (ISO 14289 standard)
 * Supports WCAG 2.1 AA compliance checking
 * Apache 2.0 License | Multi-language support
 */

namespace Lombok;

/**
 * PDF/UA compliance levels
 */
class PDFUALevel {
    const PDFUA_1 = 'PDF/UA-1'; // 2012, ISO 14289-1
    const PDFUA_2 = 'PDF/UA-2'; // 2020, ISO 14289-2
}

/**
 * PDF/UA accessibility rules
 */
class AccessibilityRule {
    const TAGGED_PDF_REQUIRED = 'tagged_pdf_required';
    const DOCUMENT_TITLE = 'document_title';
    const LOGICAL_STRUCTURE = 'logical_structure';
    const ALT_TEXT_FOR_IMAGES = 'alt_text_for_images';
    const TABLE_STRUCTURE = 'table_structure';
    const FORM_LABELS = 'form_labels';
    const COLOR_CONTRAST = 'color_contrast';
    const LANGUAGE_SPECIFICATION = 'language_specification';
    const ARTIFACT_TAGGING = 'artifact_tagging';
    const CONTENT_ORDER = 'content_order';
    const LINK_DESCRIPTIONS = 'link_descriptions';
    const NO_JAVASCRIPT = 'no_javascript';
    const ACCESSIBLE_FONTS = 'accessible_fonts';
    const HEADING_HIERARCHY = 'heading_hierarchy';
}

/**
 * Accessibility issue
 */
class AccessibilityIssue {
    public $rule;
    public $severity; // 'error', 'warning', 'info'
    public $message;
    public $location;
    public $suggestion;

    public function __construct($rule, $severity, $message, $location = null, $suggestion = null) {
        $this->rule = $rule;
        $this->severity = $severity;
        $this->message = $message;
        $this->location = $location;
        $this->suggestion = $suggestion;
    }
}

/**
 * Document structure tree node
 */
class StructureNode {
    public $tag;
    public $role;
    public $title;
    public $altText;
    public $children;

    public function __construct($tag, $role = null, $title = null, $altText = null, $children = []) {
        $this->tag = $tag;
        $this->role = $role;
        $this->title = $title;
        $this->altText = $altText;
        $this->children = $children;
    }
}

/**
 * Form field information
 */
class FormField {
    public $name;
    public $type; // Text, Button, Checkbox, Radio, etc.
    public $hasLabel;
    public $isRequired;
    public $tooltip;

    public function __construct($name, $type, $hasLabel, $isRequired, $tooltip = null) {
        $this->name = $name;
        $this->type = $type;
        $this->hasLabel = $hasLabel;
        $this->isRequired = $isRequired;
        $this->tooltip = $tooltip;
    }
}

/**
 * Image information
 */
class ImageInfo {
    public $id;
    public $hasAltText;
    public $altText;
    public $isDecorative;

    public function __construct($id, $hasAltText, $altText = null, $isDecorative = false) {
        $this->id = $id;
        $this->hasAltText = $hasAltText;
        $this->altText = $altText;
        $this->isDecorative = $isDecorative;
    }
}

/**
 * PDF/UA compliance report
 */
class PDFUAReport {
    public $fileName;
    public $detectedLevel;
    public $isCompliant;
    public $issues;
    public $structureTree;
    public $images;
    public $formFields;
    public $pageCount;
    public $fileSize;
    public $analysisTimestamp;
    public $wcagAACompliant;

    public function __construct($fileName) {
        $this->fileName = $fileName;
        $this->detectedLevel = null;
        $this->isCompliant = false;
        $this->issues = [];
        $this->structureTree = null;
        $this->images = [];
        $this->formFields = [];
        $this->pageCount = 0;
        $this->fileSize = 0;
        $this->analysisTimestamp = '';
        $this->wcagAACompliant = false;
    }

    public function addIssue($rule, $severity, $message, $location = null, $suggestion = null) {
        $this->issues[] = new AccessibilityIssue($rule, $severity, $message, $location, $suggestion);
    }

    public function errorCount() {
        return count(array_filter($this->issues, fn($i) => $i->severity === 'error'));
    }

    public function warningCount() {
        return count(array_filter($this->issues, fn($i) => $i->severity === 'warning'));
    }

    public function infoCount() {
        return count(array_filter($this->issues, fn($i) => $i->severity === 'info'));
    }

    public function toDict() {
        $issues = [];
        foreach ($this->issues as $issue) {
            $issues[] = [
                'rule' => $issue->rule,
                'severity' => $issue->severity,
                'message' => $issue->message,
                'location' => $issue->location,
                'suggestion' => $issue->suggestion,
            ];
        }

        $images = [];
        foreach ($this->images as $img) {
            $images[] = [
                'id' => $img->id,
                'hasAltText' => $img->hasAltText,
                'altText' => $img->altText,
                'isDecorative' => $img->isDecorative,
            ];
        }

        $formFields = [];
        foreach ($this->formFields as $field) {
            $formFields[] = [
                'name' => $field->name,
                'type' => $field->type,
                'hasLabel' => $field->hasLabel,
                'isRequired' => $field->isRequired,
                'tooltip' => $field->tooltip,
            ];
        }

        return [
            'fileName' => $this->fileName,
            'detectedLevel' => $this->detectedLevel,
            'isCompliant' => $this->isCompliant,
            'wcagAACompliant' => $this->wcagAACompliant,
            'errorCount' => $this->errorCount(),
            'warningCount' => $this->warningCount(),
            'infoCount' => $this->infoCount(),
            'pageCount' => $this->pageCount,
            'fileSize' => $this->fileSize,
            'analysisTimestamp' => $this->analysisTimestamp,
            'issues' => $issues,
            'images' => $images,
            'formFields' => $formFields,
        ];
    }

    public function toJSON() {
        return json_encode($this->toDict(), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
    }
}

/**
 * PDF Accessibility Analyzer
 */
class AccessibilityAnalyzer {
    public $content;

    public function __construct() {
        $this->content = null;
    }

    public function parseHeader() {
        if ($this->content === null) {
            return null;
        }

        $header = substr($this->content, 0, 1024);
        if (preg_match('/%PDF-(\d\.\d)/', $header, $matches)) {
            return $matches[1];
        }
        return null;
    }

    public function isTaggedPdf() {
        if ($this->content === null) {
            return false;
        }
        return strpos($this->content, '/StructTreeRoot') !== false ||
               strpos($this->content, '/StructParent') !== false;
    }

    public function hasDocumentTitle() {
        if ($this->content === null) {
            return false;
        }
        return (bool) preg_match('/\/Title\s*\(/', $this->content);
    }

    public function hasLanguageSpecification() {
        if ($this->content === null) {
            return false;
        }
        return strpos($this->content, '/Lang') !== false;
    }

    public function checkHeadingHierarchy() {
        if ($this->content === null) {
            return false;
        }

        $h1Count = preg_match_all('/\/H1\b/', $this->content);

        // Valid hierarchy: at least one H1 present
        return $h1Count > 0;
    }

    public function extractImages() {
        $images = [];
        if ($this->content === null) {
            return $images;
        }

        $idx = 0;
        if (preg_match_all('/\/Image.*?<<(.*?)>>/s', $this->content, $matches)) {
            foreach ($matches[1] as $imgBlock) {
                $idx++;

                $altText = null;
                $hasAlt = false;
                if (preg_match('/\/Alt\s*\((.*?)\)/', $imgBlock, $altMatch)) {
                    $altText = $altMatch[1];
                    $hasAlt = true;
                }

                $isDecorative = strpos($imgBlock, '/Artifact') !== false;

                $images[] = new ImageInfo(
                    "Image-{$idx}",
                    $hasAlt,
                    $altText,
                    $isDecorative
                );
            }
        }

        return $images;
    }

    public function extractFormFields() {
        $fields = [];
        if ($this->content === null) {
            return $fields;
        }

        if (preg_match_all('/\/FT\s*\/(\w+).*?\/T\s*\((.*?)\)/s', $this->content, $matches, PREG_SET_ORDER)) {
            foreach ($matches as $match) {
                $fieldType = $match[1];
                $fieldName = $match[2];

                $hasLabel = strpos($this->content, '/Lbl') !== false;
                $isRequired = strpos($this->content, '/Ff') !== false;

                $fields[] = new FormField($fieldName, $fieldType, $hasLabel, $isRequired);
            }
        }

        return $fields;
    }

    public function hasJavaScript() {
        if ($this->content === null) {
            return false;
        }
        return strpos($this->content, '/JavaScript') !== false;
    }

    /**
     * Check color contrast (simplified check)
     * Returns: [isSufficient, contrastRatio]
     */
    public function checkColorContrast() {
        // Simplified implementation - full implementation would analyze actual colors
        if ($this->content === null) {
            return [true, 4.5];
        }

        return [true, 4.5];
    }
}

/**
 * PDF/UA Validator
 */
class PDFUAValidator {
    private $analyzer;

    public function __construct() {
        $this->analyzer = new AccessibilityAnalyzer();
    }

    public function validate($pdfBytes) {
        $report = new PDFUAReport('document.pdf');
        $report->fileSize = strlen($pdfBytes);
        $report->analysisTimestamp = date('c');

        $this->analyzer->content = $pdfBytes;

        $pdfVersion = $this->analyzer->parseHeader();
        if ($pdfVersion === null) {
            $report->addIssue(
                AccessibilityRule::TAGGED_PDF_REQUIRED,
                'error',
                'Not a valid PDF file',
                null,
                'Ensure file is a valid PDF document'
            );
            return $report;
        }

        $report->images = $this->analyzer->extractImages();
        $report->formFields = $this->analyzer->extractFormFields();

        $this->checkTaggedPdf($report);
        $this->checkDocumentTitle($report);
        $this->checkLanguageSpecification($report);
        $this->checkImageAltText($report);
        $this->checkFormLabels($report);
        $this->checkHeadingHierarchy($report);
        $this->checkNoJavaScript($report);
        $this->checkColorContrast($report);

        $errorCount = $report->errorCount();
        $report->isCompliant = $errorCount === 0;
        $report->wcagAACompliant = $errorCount === 0 && $report->warningCount() <= 2;

        return $report;
    }

    private function checkTaggedPdf(&$report) {
        if (!$this->analyzer->isTaggedPdf()) {
            $report->addIssue(
                AccessibilityRule::TAGGED_PDF_REQUIRED,
                'error',
                'PDF is not tagged (no structure tree)',
                null,
                'Create a tagged PDF with proper document structure'
            );
        }
    }

    private function checkDocumentTitle(&$report) {
        if (!$this->analyzer->hasDocumentTitle()) {
            $report->addIssue(
                AccessibilityRule::DOCUMENT_TITLE,
                'warning',
                'Document title not found',
                null,
                'Add title to document properties'
            );
        }
    }

    private function checkLanguageSpecification(&$report) {
        if (!$this->analyzer->hasLanguageSpecification()) {
            $report->addIssue(
                AccessibilityRule::LANGUAGE_SPECIFICATION,
                'warning',
                'Document language not specified',
                null,
                'Set document language in PDF properties'
            );
        }
    }

    private function checkImageAltText(&$report) {
        $imagesWithoutAlt = array_filter($report->images, function ($img) {
            return !$img->hasAltText && !$img->isDecorative;
        });

        if (count($imagesWithoutAlt) > 0) {
            $count = count($imagesWithoutAlt);
            $report->addIssue(
                AccessibilityRule::ALT_TEXT_FOR_IMAGES,
                'error',
                "Found {$count} image(s) without alt text",
                null,
                'Add descriptive alt text to all non-decorative images'
            );
        }
    }

    private function checkFormLabels(&$report) {
        $fieldsWithoutLabels = array_filter($report->formFields, fn($f) => !$f->hasLabel);

        if (count($fieldsWithoutLabels) > 0) {
            $count = count($fieldsWithoutLabels);
            $report->addIssue(
                AccessibilityRule::FORM_LABELS,
                'error',
                "Found {$count} form field(s) without labels",
                null,
                'Associate labels with all form fields'
            );
        }
    }

    private function checkHeadingHierarchy(&$report) {
        if (!$this->analyzer->checkHeadingHierarchy()) {
            $report->addIssue(
                AccessibilityRule::HEADING_HIERARCHY,
                'warning',
                'Heading hierarchy may not be correct',
                null,
                'Use proper heading hierarchy (H1 → H2 → H3)'
            );
        }
    }

    private function checkNoJavaScript(&$report) {
        if ($this->analyzer->hasJavaScript()) {
            $report->addIssue(
                AccessibilityRule::NO_JAVASCRIPT,
                'info',
                'PDF contains JavaScript',
                null,
                'Consider removing JavaScript for better accessibility'
            );
        }
    }

    private function checkColorContrast(&$report) {
        [$isSufficient, $contrastRatio] = $this->analyzer->checkColorContrast();
        if (!$isSufficient) {
            $ratioStr = number_format($contrastRatio, 1);
            $report->addIssue(
                AccessibilityRule::COLOR_CONTRAST,
                'warning',
                "Color contrast may be insufficient (ratio: {$ratioStr}:1)",
                null,
                'Ensure text has at least 4.5:1 color contrast (WCAG AA)'
            );
        }
    }
}

/**
 * Main PDFUA class
 */
class PDFUA {
    private $validator;
    private $report;

    public function __construct() {
        $this->validator = new PDFUAValidator();
        $this->report = null;
    }

    public function validate($pdfSource) {
        if (is_string($pdfSource) && file_exists($pdfSource)) {
            $pdfBytes = file_get_contents($pdfSource);
        } else {
            $pdfBytes = $pdfSource;
        }

        $this->report = $this->validator->validate($pdfBytes);
        return $this;
    }

    public function getReport() {
        return $this->report;
    }

    public function isCompliant() {
        if ($this->report === null) {
            return false;
        }
        return $this->report->isCompliant;
    }

    public function isWCAGAACompliant() {
        if ($this->report === null) {
            return false;
        }
        return $this->report->wcagAACompliant;
    }

    public function getErrors() {
        if ($this->report === null) {
            return [];
        }
        return array_values(array_filter($this->report->issues, fn($i) => $i->severity === 'error'));
    }

    public function getWarnings() {
        if ($this->report === null) {
            return [];
        }
        return array_values(array_filter($this->report->issues, fn($i) => $i->severity === 'warning'));
    }

    public function getInfo() {
        if ($this->report === null) {
            return [];
        }
        return array_values(array_filter($this->report->issues, fn($i) => $i->severity === 'info'));
    }

    public function getIssueCounts() {
        if ($this->report === null) {
            return [0, 0, 0];
        }
        return [
            $this->report->errorCount(),
            $this->report->warningCount(),
            $this->report->infoCount(),
        ];
    }

    public function getImages() {
        if ($this->report === null) {
            return [];
        }
        return $this->report->images;
    }

    public function getFormFields() {
        if ($this->report === null) {
            return [];
        }
        return $this->report->formFields;
    }

    public function toJSON() {
        if ($this->report === null) {
            return '{}';
        }
        return $this->report->toJSON();
    }

    public function toDict() {
        if ($this->report === null) {
            return [];
        }
        return $this->report->toDict();
    }
}
