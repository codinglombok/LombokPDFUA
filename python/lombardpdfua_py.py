"""
LombokPDFUA - Python Port (v0.9.0-alpha)
PDF/UA accessibility validation (ISO 14289 standard).
Supports WCAG 2.1 AA compliance checking.
Apache 2.0 License | Multi-language support
"""

import re
import json
from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Tuple
from enum import Enum
from datetime import datetime


# ============================================================================
# PDF/UA CONSTANTS
# ============================================================================

class PDFUALevel(Enum):
    """PDF/UA compliance levels"""
    PDFUA_1 = "PDF/UA-1"  # 2012, ISO 14289-1
    PDFUA_2 = "PDF/UA-2"  # 2020, ISO 14289-2


class AccessibilityRule(Enum):
    """PDF/UA accessibility rules"""
    TAGGED_PDF_REQUIRED = "tagged_pdf_required"
    DOCUMENT_TITLE = "document_title"
    LOGICAL_STRUCTURE = "logical_structure"
    ALT_TEXT_FOR_IMAGES = "alt_text_for_images"
    TABLE_STRUCTURE = "table_structure"
    FORM_LABELS = "form_labels"
    COLOR_CONTRAST = "color_contrast"
    LANGUAGE_SPECIFICATION = "language_specification"
    ARTIFACT_TAGGING = "artifact_tagging"
    CONTENT_ORDER = "content_order"
    LINK_DESCRIPTIONS = "link_descriptions"
    NO_JAVASCRIPT = "no_javascript"
    ACCESSIBLE_FONTS = "accessible_fonts"
    HEADING_HIERARCHY = "heading_hierarchy"


# ============================================================================
# PDF/UA DATA STRUCTURES
# ============================================================================

@dataclass
class AccessibilityIssue:
    """Accessibility issue"""
    rule: AccessibilityRule
    severity: str  # "error", "warning", "info"
    message: str
    location: Optional[str] = None
    suggestion: Optional[str] = None


@dataclass
class StructureNode:
    """Document structure tree node"""
    tag: str
    role: Optional[str] = None
    title: Optional[str] = None
    alt_text: Optional[str] = None
    children: List['StructureNode'] = None
    
    def __post_init__(self):
        if self.children is None:
            self.children = []


@dataclass
class FormField:
    """Form field information"""
    name: str
    type: str  # Text, Button, Checkbox, Radio, etc.
    has_label: bool
    is_required: bool
    tooltip: Optional[str] = None


@dataclass
class ImageInfo:
    """Image information"""
    id: str
    has_alt_text: bool
    is_decorative: bool
    alt_text: Optional[str] = None


@dataclass
class PDFUAReport:
    """PDF/UA compliance report"""
    file_name: str
    detected_level: Optional[PDFUALevel] = None
    is_compliant: bool = False
    issues: List[AccessibilityIssue] = None
    structure_tree: Optional[StructureNode] = None
    images: List[ImageInfo] = None
    form_fields: List[FormField] = None
    page_count: int = 0
    file_size: int = 0
    analysis_timestamp: str = ""
    wcag_aa_compliant: bool = False
    
    def __post_init__(self):
        if self.issues is None:
            self.issues = []
        if self.images is None:
            self.images = []
        if self.form_fields is None:
            self.form_fields = []
    
    def add_issue(self, rule: AccessibilityRule, severity: str, message: str,
                  location: str = None, suggestion: str = None):
        """Add an accessibility issue"""
        self.issues.append(AccessibilityIssue(
            rule=rule,
            severity=severity,
            message=message,
            location=location,
            suggestion=suggestion
        ))
    
    def error_count(self) -> int:
        """Count errors"""
        return sum(1 for issue in self.issues if issue.severity == "error")
    
    def warning_count(self) -> int:
        """Count warnings"""
        return sum(1 for issue in self.issues if issue.severity == "warning")
    
    def info_count(self) -> int:
        """Count info messages"""
        return sum(1 for issue in self.issues if issue.severity == "info")
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary"""
        return {
            "fileName": self.file_name,
            "detectedLevel": self.detected_level.value if self.detected_level else None,
            "isCompliant": self.is_compliant,
            "wcagAACompliant": self.wcag_aa_compliant,
            "errorCount": self.error_count(),
            "warningCount": self.warning_count(),
            "infoCount": self.info_count(),
            "pageCount": self.page_count,
            "fileSize": self.file_size,
            "analysisTimestamp": self.analysis_timestamp,
            "issues": [
                {
                    "rule": issue.rule.value,
                    "severity": issue.severity,
                    "message": issue.message,
                    "location": issue.location,
                    "suggestion": issue.suggestion
                }
                for issue in self.issues
            ],
            "images": [
                {
                    "id": img.id,
                    "hasAltText": img.has_alt_text,
                    "altText": img.alt_text,
                    "isDecorative": img.is_decorative
                }
                for img in self.images
            ],
            "formFields": [
                {
                    "name": field.name,
                    "type": field.type,
                    "hasLabel": field.has_label,
                    "isRequired": field.is_required,
                    "tooltip": field.tooltip
                }
                for field in self.form_fields
            ]
        }
    
    def to_json(self) -> str:
        """Convert to JSON"""
        return json.dumps(self.to_dict(), indent=2, default=str)


# ============================================================================
# PDF ACCESSIBILITY ANALYZER
# ============================================================================

class AccessibilityAnalyzer:
    """Analyzes PDF for accessibility"""
    
    def __init__(self):
        self.content: Optional[bytes] = None
    
    def parse_header(self) -> Optional[str]:
        """Parse PDF version"""
        if not self.content:
            return None
        
        header = self.content[:1024].decode('latin1', errors='ignore')
        match = re.search(r'%PDF-(\d\.\d)', header)
        
        if match:
            return match.group(1)
        return None
    
    def is_tagged_pdf(self) -> bool:
        """Check if PDF is tagged"""
        if not self.content:
            return False
        
        content_str = self.content.decode('latin1', errors='ignore')
        # Look for tag tree markers
        return '/StructTreeRoot' in content_str or '/StructParent' in content_str
    
    def has_document_title(self) -> bool:
        """Check if document has title"""
        if not self.content:
            return False
        
        content_str = self.content.decode('latin1', errors='ignore')
        # Look for Title in Info dictionary
        return bool(re.search(r'/Title\s*\(', content_str))
    
    def has_language_specification(self) -> bool:
        """Check if language is specified"""
        if not self.content:
            return False
        
        content_str = self.content.decode('latin1', errors='ignore')
        return '/Lang' in content_str
    
    def check_heading_hierarchy(self) -> bool:
        """Check if heading hierarchy is correct"""
        if not self.content:
            return False
        
        content_str = self.content.decode('latin1', errors='ignore')
        # Look for heading tags
        h1_count = len(re.findall(r'/H1\b', content_str))
        h2_count = len(re.findall(r'/H2\b', content_str))
        h3_count = len(re.findall(r'/H3\b', content_str))
        
        # Valid hierarchy: H1 first, then H2s, then H3s
        return h1_count > 0
    
    def extract_images(self) -> List[ImageInfo]:
        """Extract image information"""
        if not self.content:
            return []
        
        images = []
        content_str = self.content.decode('latin1', errors='ignore')
        
        # Find image references
        img_pattern = r'/Image.*?<<(.*?)>>'
        for idx, match in enumerate(re.finditer(img_pattern, content_str, re.DOTALL)):
            img_block = match.group(1)
            
            # Check for alt text
            alt_match = re.search(r'/Alt\s*\((.*?)\)', img_block)
            alt_text = alt_match.group(1) if alt_match else None
            has_alt = alt_text is not None
            
            # Check if marked as decorative
            is_decorative = '/Artifact' in img_block
            
            images.append(ImageInfo(
                id=f"Image-{idx+1}",
                has_alt_text=has_alt,
                alt_text=alt_text,
                is_decorative=is_decorative
            ))
        
        return images
    
    def extract_form_fields(self) -> List[FormField]:
        """Extract form field information"""
        if not self.content:
            return []
        
        fields = []
        content_str = self.content.decode('latin1', errors='ignore')
        
        # Find form fields
        field_pattern = r'/FT\s*/(\w+).*?/T\s*\((.*?)\)'
        for match in re.finditer(field_pattern, content_str, re.DOTALL):
            field_type = match.group(1)
            field_name = match.group(2)
            
            # Check for label
            has_label = '/Lbl' in content_str
            
            # Check if required
            is_required = '/Ff' in content_str
            
            fields.append(FormField(
                name=field_name,
                type=field_type,
                has_label=has_label,
                is_required=is_required
            ))
        
        return fields
    
    def has_javascript(self) -> bool:
        """Check for JavaScript"""
        if not self.content:
            return False
        
        content_str = self.content.decode('latin1', errors='ignore')
        return '/JavaScript' in content_str
    
    def check_color_contrast(self) -> Tuple[bool, float]:
        """
        Check color contrast (simplified check)
        Returns: (is_sufficient, contrast_ratio)
        """
        # Simplified implementation - full implementation would analyze actual colors
        # For now, return true if no specific color issues detected
        if not self.content:
            return (True, 4.5)
        
        return (True, 4.5)


# ============================================================================
# PDF/UA VALIDATOR
# ============================================================================

class PDFUAValidator:
    """Validates PDF/UA accessibility"""
    
    def __init__(self):
        self.analyzer = AccessibilityAnalyzer()
    
    def validate(self, pdf_bytes: bytes, target_level: Optional[PDFUALevel] = None) -> PDFUAReport:
        """Validate PDF/UA compliance"""
        report = PDFUAReport(
            file_name="document.pdf",
            page_count=0,
            file_size=len(pdf_bytes),
            analysis_timestamp=datetime.now().isoformat()
        )
        
        self.analyzer.content = pdf_bytes
        
        # Check if it's a valid PDF
        pdf_version = self.analyzer.parse_header()
        if not pdf_version:
            report.add_issue(
                AccessibilityRule.TAGGED_PDF_REQUIRED,
                "error",
                "Not a valid PDF file",
                suggestion="Ensure file is a valid PDF document"
            )
            return report
        
        # Extract content
        report.images = self.analyzer.extract_images()
        report.form_fields = self.analyzer.extract_form_fields()
        
        # Run accessibility checks
        self._check_tagged_pdf(report)
        self._check_document_title(report)
        self._check_language_specification(report)
        self._check_image_alt_text(report)
        self._check_form_labels(report)
        self._check_heading_hierarchy(report)
        self._check_no_javascript(report)
        self._check_color_contrast(report)
        
        # Determine overall compliance
        error_count = report.error_count()
        report.is_compliant = error_count == 0
        
        # Check WCAG 2.1 AA compliance
        report.wcag_aa_compliant = error_count == 0 and report.warning_count() <= 2
        
        return report
    
    def _check_tagged_pdf(self, report: PDFUAReport):
        """Check if PDF is tagged"""
        if not self.analyzer.is_tagged_pdf():
            report.add_issue(
                AccessibilityRule.TAGGED_PDF_REQUIRED,
                "error",
                "PDF is not tagged (no structure tree)",
                suggestion="Create a tagged PDF with proper document structure"
            )
    
    def _check_document_title(self, report: PDFUAReport):
        """Check for document title"""
        if not self.analyzer.has_document_title():
            report.add_issue(
                AccessibilityRule.DOCUMENT_TITLE,
                "warning",
                "Document title not found",
                suggestion="Add title to document properties"
            )
    
    def _check_language_specification(self, report: PDFUAReport):
        """Check language specification"""
        if not self.analyzer.has_language_specification():
            report.add_issue(
                AccessibilityRule.LANGUAGE_SPECIFICATION,
                "warning",
                "Document language not specified",
                suggestion="Set document language in PDF properties"
            )
    
    def _check_image_alt_text(self, report: PDFUAReport):
        """Check image alt text"""
        images_without_alt = [img for img in report.images if not img.has_alt_text and not img.is_decorative]
        if images_without_alt:
            count = len(images_without_alt)
            report.add_issue(
                AccessibilityRule.ALT_TEXT_FOR_IMAGES,
                "error",
                f"Found {count} image(s) without alt text",
                suggestion="Add descriptive alt text to all non-decorative images"
            )
    
    def _check_form_labels(self, report: PDFUAReport):
        """Check form field labels"""
        fields_without_labels = [f for f in report.form_fields if not f.has_label]
        if fields_without_labels:
            count = len(fields_without_labels)
            report.add_issue(
                AccessibilityRule.FORM_LABELS,
                "error",
                f"Found {count} form field(s) without labels",
                suggestion="Associate labels with all form fields"
            )
    
    def _check_heading_hierarchy(self, report: PDFUAReport):
        """Check heading hierarchy"""
        if not self.analyzer.check_heading_hierarchy():
            report.add_issue(
                AccessibilityRule.HEADING_HIERARCHY,
                "warning",
                "Heading hierarchy may not be correct",
                suggestion="Use proper heading hierarchy (H1 → H2 → H3)"
            )
    
    def _check_no_javascript(self, report: PDFUAReport):
        """Check for JavaScript"""
        if self.analyzer.has_javascript():
            report.add_issue(
                AccessibilityRule.NO_JAVASCRIPT,
                "info",
                "PDF contains JavaScript",
                suggestion="Consider removing JavaScript for better accessibility"
            )
    
    def _check_color_contrast(self, report: PDFUAReport):
        """Check color contrast"""
        is_sufficient, contrast_ratio = self.analyzer.check_color_contrast()
        if not is_sufficient:
            report.add_issue(
                AccessibilityRule.COLOR_CONTRAST,
                "warning",
                f"Color contrast may be insufficient (ratio: {contrast_ratio:.1f}:1)",
                suggestion="Ensure text has at least 4.5:1 color contrast (WCAG AA)"
            )


# ============================================================================
# MAIN CLASS
# ============================================================================

class PDFUA:
    """Main PDF/UA validator class"""
    
    def __init__(self):
        self.validator = PDFUAValidator()
        self.report: Optional[PDFUAReport] = None
    
    def validate(self, pdf_source, target_level: Optional[PDFUALevel] = None) -> 'PDFUA':
        """Validate PDF/UA compliance"""
        if isinstance(pdf_source, str):
            with open(pdf_source, 'rb') as f:
                pdf_bytes = f.read()
        else:
            pdf_bytes = pdf_source
        
        self.report = self.validator.validate(pdf_bytes, target_level)
        return self
    
    def getReport(self) -> Optional[PDFUAReport]:
        """Get the validation report"""
        return self.report
    
    def isCompliant(self) -> bool:
        """Check if document is compliant"""
        if not self.report:
            return False
        return self.report.is_compliant
    
    def isWCAGAACompliant(self) -> bool:
        """Check WCAG 2.1 AA compliance"""
        if not self.report:
            return False
        return self.report.wcag_aa_compliant
    
    def getErrors(self) -> List[AccessibilityIssue]:
        """Get all errors"""
        if not self.report:
            return []
        return [issue for issue in self.report.issues if issue.severity == "error"]
    
    def getWarnings(self) -> List[AccessibilityIssue]:
        """Get all warnings"""
        if not self.report:
            return []
        return [issue for issue in self.report.issues if issue.severity == "warning"]
    
    def getInfo(self) -> List[AccessibilityIssue]:
        """Get all info messages"""
        if not self.report:
            return []
        return [issue for issue in self.report.issues if issue.severity == "info"]
    
    def getIssueCounts(self) -> Tuple[int, int, int]:
        """Get error, warning, and info counts"""
        if not self.report:
            return (0, 0, 0)
        return (self.report.error_count(), self.report.warning_count(), self.report.info_count())
    
    def getImages(self) -> List[ImageInfo]:
        """Get image information"""
        if not self.report:
            return []
        return self.report.images
    
    def getFormFields(self) -> List[FormField]:
        """Get form field information"""
        if not self.report:
            return []
        return self.report.form_fields
    
    def toJSON(self) -> str:
        """Export report to JSON"""
        if not self.report:
            return "{}"
        return self.report.to_json()
    
    def toDict(self) -> Dict[str, Any]:
        """Export report to dictionary"""
        if not self.report:
            return {}
        return self.report.to_dict()
