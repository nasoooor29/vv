import { format } from "date-fns";

export type ExportFormat = "csv" | "json" | "xml";

interface ExportOptions {
  filename?: string;
  headers?: string[];
  rowMapper?: (item: Record<string, unknown>) => (string | number | boolean | null | undefined)[];
}

/**
 * Generic export function that handles CSV, JSON, and XML formats
 */
export function exportData<T extends Record<string, unknown>>(
  data: T[],
  exportFormat: ExportFormat,
  options: ExportOptions = {}
): void {
  const { filename = "export", headers, rowMapper } = options;

  let content: string;
  let mimeType: string;
  let fileExtension: string;

  switch (exportFormat) {
    case "csv": {
      content = generateCSV(data, headers, rowMapper);
      mimeType = "text/csv";
      fileExtension = "csv";
      break;
    }
    case "json": {
      content = JSON.stringify(data, null, 2);
      mimeType = "application/json";
      fileExtension = "json";
      break;
    }
    case "xml": {
      content = generateXML(data, "items", "item");
      mimeType = "application/xml";
      fileExtension = "xml";
      break;
    }
  }

  downloadFile(content, `${filename}-${format(new Date(), "yyyy-MM-dd-HHmmss")}.${fileExtension}`, mimeType);
}

/**
 * Generate CSV content from data array
 */
function generateCSV<T extends Record<string, unknown>>(
  data: T[],
  headers?: string[],
  rowMapper?: (item: T) => (string | number | boolean | null | undefined)[]
): string {
  if (data.length === 0) return "";

  const csvHeaders = headers || Object.keys(data[0]);
  
  const rows = data.map((item) => {
    if (rowMapper) {
      return rowMapper(item).map(escapeCSVField);
    }
    return csvHeaders.map((header) => escapeCSVField(item[header]));
  });

  return [csvHeaders.join(","), ...rows.map((row) => row.join(","))].join("\n");
}

/**
 * Escape a field for CSV format
 */
function escapeCSVField(value: unknown): string {
  if (value === null || value === undefined) return "";
  const str = String(value);
  if (str.includes(",") || str.includes('"') || str.includes("\n")) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

/**
 * Generate XML content from data array
 */
function generateXML<T extends Record<string, unknown>>(
  data: T[],
  rootElement: string,
  itemElement: string
): string {
  const xmlItems = data
    .map((item) => {
      const fields = Object.entries(item)
        .map(([key, value]) => `    <${key}>${escapeXML(value)}</${key}>`)
        .join("\n");
      return `  <${itemElement}>\n${fields}\n  </${itemElement}>`;
    })
    .join("\n");

  return `<?xml version="1.0" encoding="UTF-8"?>\n<${rootElement}>\n${xmlItems}\n</${rootElement}>`;
}

/**
 * Escape special characters for XML
 */
function escapeXML(value: unknown): string {
  if (value === null || value === undefined) return "";
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;");
}

/**
 * Trigger a file download in the browser
 */
function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
