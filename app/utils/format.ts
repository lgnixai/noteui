export function formatFieldValue(value: any, type: string): string {
  if (value === undefined || value === null) {
    return '';
  }

  switch (type) {
    case 'text':
      return String(value);
    case 'number':
      return Number(value).toLocaleString();
    case 'boolean':
      return value ? 'Yes' : 'No';
    case 'date':
      return new Date(value).toLocaleDateString();
    default:
      return String(value);
  }
} 