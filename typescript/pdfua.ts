/**
 * LombokPDFUA - TypeScript Port (v0.9.0-alpha)
 * PDF/UA accessibility validation (ISO 14289 standard).
 * Supports WCAG 2.1 AA compliance checking.
 * Apache 2.0 License | Multi-language support
 */

// ============================================================================
// PDF/UA CONSTANTS
// ============================================================================

export enum PDFUALevel {
  PDFUA_1 = 'PDF/UA-1', // 2012, ISO 14289-1
  PDFUA_2 = 'PDF/UA-2', // 2020, ISO 14289-2
}

export enum AccessibilityRule {
  TAGGED_PDF_REQUIRED = 'tagged_pdf_required',
  DOCUMENT_TITLE = 'document_title',
  LOGICAL_STRUCTURE = 'logical_structure',
  ALT_TEXT_FOR_IMAGES = 'alt_text_for_images',
  TABLE_STRUCTURE = 'table_structure',
  FORM_LABELS = 'form_labels',
  COLOR_CONTRAST = 'color_contrast',
  LANGUAGE_SPECIFICATION = 'language_specification',
  ARTIFACT_TAGGING = 'artifact_tagging',
  CONTENT_ORDER = 'content_order',
  LINK_DESCRIPTIONS = 'link_descriptions',
  NO_JAVASCRIPT = 'no_javascript',
  ACCESSIBLE_FONTS = 'accessible_fonts',
  HEADING_HIERARCHY = 'heading_hierarchy',
}

// ============================================================================
// PDF/UA DATA STRUCTURES
// ============================================================================

export interface AccessibilityIssue {
  rule: AccessibilityRule;
  severity: 'error' | 'warning' | 'info';
  message: string;
  location?: string;
  suggestion?: string;
}

export interface StructureNode {
  tag: string;
  role?: string;
  title?: string;
  altText?: string;
  children: StructureNode[];
}

export interface FormField {
  name: string;
  type: string; // Text, Button, Checkbox, Radio, etc.
  hasLabel: boolean;
  isRequired: boolean;
  tooltip?: string;
}

export interface ImageInfo {
  id: string;
  hasAltText: boolean;
  altText?: string;
  isDecorative: boolean;
}

export class PDFUAReport {
  fileName: string;
  detectedLevel: PDFUALevel | null = null;
  isCompliant = false;
  issues: AccessibilityIssue[] = [];
  structureTree: StructureNode | null = null;
  images: ImageInfo[] = [];
  formFields: FormField[] = [];
  pageCount = 0;
  fileSize = 0;
  analysisTimestamp = '';
  wcagAACompliant = false;

  constructor(fileName: string) {
    this.fileName = fileName;
  }

  addIssue(
    rule: AccessibilityRule,
    severity: 'error' | 'warning' | 'info',
    message: string,
    location?: string,
    suggestion?: string
  ): void {
    this.issues.push({ rule, severity, message, location, suggestion });
  }

  errorCount(): number {
    return this.issues.filter((i) => i.severity === 'error').length;
  }

  warningCount(): number {
    return this.issues.filter((i) => i.severity === 'warning').length;
  }

  infoCount(): number {
    return this.issues.filter((i) => i.severity === 'info').length;
  }

  toDict(): Record<string, unknown> {
    return {
      fileName: this.fileName,
      detectedLevel: this.detectedLevel,
      isCompliant: this.isCompliant,
      wcagAACompliant: this.wcagAACompliant,
      errorCount: this.errorCount(),
      warningCount: this.warningCount(),
      infoCount: this.infoCount(),
      pageCount: this.pageCount,
      fileSize: this.fileSize,
      analysisTimestamp: this.analysisTimestamp,
      issues: this.issues,
      images: this.images,
      formFields: this.formFields,
    };
  }

  toJSON(): string {
    return JSON.stringify(this.toDict(), null, 2);
  }
}

// ============================================================================
// PDF ACCESSIBILITY ANALYZER
// ============================================================================

export class AccessibilityAnalyzer {
  content: Uint8Array | null = null;

  private toStr(): string {
    if (!this.content) return '';
    return Array.from(this.content)
      .map((b) => String.fromCharCode(b))
      .join('');
  }

  parseHeader(): string | null {
    if (!this.content) return null;
    const end = Math.min(1024, this.content.length);
    const header = Array.from(this.content.slice(0, end))
      .map((b) => String.fromCharCode(b))
      .join('');
    const match = header.match(/%PDF-(\d\.\d)/);
    return match ? match[1] : null;
  }

  isTaggedPdf(): boolean {
    if (!this.content) return false;
    const content = this.toStr();
    return content.includes('/StructTreeRoot') || content.includes('/StructParent');
  }

  hasDocumentTitle(): boolean {
    if (!this.content) return false;
    return /\/Title\s*\(/.test(this.toStr());
  }

  hasLanguageSpecification(): boolean {
    if (!this.content) return false;
    return this.toStr().includes('/Lang');
  }

  checkHeadingHierarchy(): boolean {
    if (!this.content) return false;
    const content = this.toStr();
    const h1Matches = content.match(/\/H1\b/g);
    // Valid hierarchy: at least one H1 present
    return !!h1Matches && h1Matches.length > 0;
  }

  extractImages(): ImageInfo[] {
    const images: ImageInfo[] = [];
    if (!this.content) return images;

    const content = this.toStr();
    const imgPattern = /\/Image.*?<<(.*?)>>/gs;
    let match: RegExpExecArray | null;
    let idx = 0;

    while ((match = imgPattern.exec(content)) !== null) {
      idx++;
      const imgBlock = match[1];

      const altMatch = imgBlock.match(/\/Alt\s*\((.*?)\)/);
      const altText = altMatch ? altMatch[1] : undefined;
      const hasAlt = altText !== undefined;

      const isDecorative = imgBlock.includes('/Artifact');

      images.push({
        id: `Image-${idx}`,
        hasAltText: hasAlt,
        altText,
        isDecorative,
      });
    }

    return images;
  }

  extractFormFields(): FormField[] {
    const fields: FormField[] = [];
    if (!this.content) return fields;

    const content = this.toStr();
    const fieldPattern = /\/FT\s*\/(\w+).*?\/T\s*\((.*?)\)/gs;
    let match: RegExpExecArray | null;

    const hasLabel = content.includes('/Lbl');
    const isRequired = content.includes('/Ff');

    while ((match = fieldPattern.exec(content)) !== null) {
      const fieldType = match[1];
      const fieldName = match[2];

      fields.push({
        name: fieldName,
        type: fieldType,
        hasLabel,
        isRequired,
      });
    }

    return fields;
  }

  hasJavaScript(): boolean {
    if (!this.content) return false;
    return this.toStr().includes('/JavaScript');
  }

  /**
   * Check color contrast (simplified check)
   * Returns: [isSufficient, contrastRatio]
   */
  checkColorContrast(): [boolean, number] {
    // Simplified implementation - full implementation would analyze actual colors
    if (!this.content) return [true, 4.5];
    return [true, 4.5];
  }
}

// ============================================================================
// PDF/UA VALIDATOR
// ============================================================================

export class PDFUAValidator {
  analyzer: AccessibilityAnalyzer;

  constructor() {
    this.analyzer = new AccessibilityAnalyzer();
  }

  validate(pdfBytes: Uint8Array): PDFUAReport {
    const report = new PDFUAReport('document.pdf');
    report.fileSize = pdfBytes.length;
    report.analysisTimestamp = new Date().toISOString();

    this.analyzer.content = pdfBytes;

    const pdfVersion = this.analyzer.parseHeader();
    if (!pdfVersion) {
      report.addIssue(
        AccessibilityRule.TAGGED_PDF_REQUIRED,
        'error',
        'Not a valid PDF file',
        undefined,
        'Ensure file is a valid PDF document'
      );
      return report;
    }

    report.images = this.analyzer.extractImages();
    report.formFields = this.analyzer.extractFormFields();

    this.checkTaggedPdf(report);
    this.checkDocumentTitle(report);
    this.checkLanguageSpecification(report);
    this.checkImageAltText(report);
    this.checkFormLabels(report);
    this.checkHeadingHierarchy(report);
    this.checkNoJavaScript(report);
    this.checkColorContrast(report);

    const errorCount = report.errorCount();
    report.isCompliant = errorCount === 0;
    report.wcagAACompliant = errorCount === 0 && report.warningCount() <= 2;

    return report;
  }

  private checkTaggedPdf(report: PDFUAReport): void {
    if (!this.analyzer.isTaggedPdf()) {
      report.addIssue(
        AccessibilityRule.TAGGED_PDF_REQUIRED,
        'error',
        'PDF is not tagged (no structure tree)',
        undefined,
        'Create a tagged PDF with proper document structure'
      );
    }
  }

  private checkDocumentTitle(report: PDFUAReport): void {
    if (!this.analyzer.hasDocumentTitle()) {
      report.addIssue(
        AccessibilityRule.DOCUMENT_TITLE,
        'warning',
        'Document title not found',
        undefined,
        'Add title to document properties'
      );
    }
  }

  private checkLanguageSpecification(report: PDFUAReport): void {
    if (!this.analyzer.hasLanguageSpecification()) {
      report.addIssue(
        AccessibilityRule.LANGUAGE_SPECIFICATION,
        'warning',
        'Document language not specified',
        undefined,
        'Set document language in PDF properties'
      );
    }
  }

  private checkImageAltText(report: PDFUAReport): void {
    const imagesWithoutAlt = report.images.filter((img) => !img.hasAltText && !img.isDecorative);
    if (imagesWithoutAlt.length > 0) {
      const count = imagesWithoutAlt.length;
      report.addIssue(
        AccessibilityRule.ALT_TEXT_FOR_IMAGES,
        'error',
        `Found ${count} image(s) without alt text`,
        undefined,
        'Add descriptive alt text to all non-decorative images'
      );
    }
  }

  private checkFormLabels(report: PDFUAReport): void {
    const fieldsWithoutLabels = report.formFields.filter((f) => !f.hasLabel);
    if (fieldsWithoutLabels.length > 0) {
      const count = fieldsWithoutLabels.length;
      report.addIssue(
        AccessibilityRule.FORM_LABELS,
        'error',
        `Found ${count} form field(s) without labels`,
        undefined,
        'Associate labels with all form fields'
      );
    }
  }

  private checkHeadingHierarchy(report: PDFUAReport): void {
    if (!this.analyzer.checkHeadingHierarchy()) {
      report.addIssue(
        AccessibilityRule.HEADING_HIERARCHY,
        'warning',
        'Heading hierarchy may not be correct',
        undefined,
        'Use proper heading hierarchy (H1 -> H2 -> H3)'
      );
    }
  }

  private checkNoJavaScript(report: PDFUAReport): void {
    if (this.analyzer.hasJavaScript()) {
      report.addIssue(
        AccessibilityRule.NO_JAVASCRIPT,
        'info',
        'PDF contains JavaScript',
        undefined,
        'Consider removing JavaScript for better accessibility'
      );
    }
  }

  private checkColorContrast(report: PDFUAReport): void {
    const [isSufficient, contrastRatio] = this.analyzer.checkColorContrast();
    if (!isSufficient) {
      report.addIssue(
        AccessibilityRule.COLOR_CONTRAST,
        'warning',
        `Color contrast may be insufficient (ratio: ${contrastRatio.toFixed(1)}:1)`,
        undefined,
        'Ensure text has at least 4.5:1 color contrast (WCAG AA)'
      );
    }
  }
}

// ============================================================================
// MAIN CLASS
// ============================================================================

export class PDFUA {
  private validator: PDFUAValidator;
  private report: PDFUAReport | null = null;

  constructor() {
    this.validator = new PDFUAValidator();
  }

  validate(pdfBytes: Uint8Array): PDFUA {
    this.report = this.validator.validate(pdfBytes);
    return this;
  }

  getReport(): PDFUAReport | null {
    return this.report;
  }

  isCompliant(): boolean {
    if (!this.report) return false;
    return this.report.isCompliant;
  }

  isWCAGAACompliant(): boolean {
    if (!this.report) return false;
    return this.report.wcagAACompliant;
  }

  getErrors(): AccessibilityIssue[] {
    if (!this.report) return [];
    return this.report.issues.filter((i) => i.severity === 'error');
  }

  getWarnings(): AccessibilityIssue[] {
    if (!this.report) return [];
    return this.report.issues.filter((i) => i.severity === 'warning');
  }

  getInfo(): AccessibilityIssue[] {
    if (!this.report) return [];
    return this.report.issues.filter((i) => i.severity === 'info');
  }

  getIssueCounts(): [number, number, number] {
    if (!this.report) return [0, 0, 0];
    return [this.report.errorCount(), this.report.warningCount(), this.report.infoCount()];
  }

  getImages(): ImageInfo[] {
    if (!this.report) return [];
    return this.report.images;
  }

  getFormFields(): FormField[] {
    if (!this.report) return [];
    return this.report.formFields;
  }

  toJSON(): string {
    if (!this.report) return '{}';
    return this.report.toJSON();
  }

  toDict(): Record<string, unknown> {
    if (!this.report) return {};
    return this.report.toDict();
  }
}

export default PDFUA;
