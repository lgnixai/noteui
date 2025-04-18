import React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Field, Record, RecordFormData } from '../types/record';
import { validateRecord } from '../types/record';

interface RecordDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: RecordFormData) => void;
  fields: Field[];
  record?: Record;
}

export function RecordDialog({ isOpen, onClose, onSubmit, fields, record }: RecordDialogProps) {
  const [formData, setFormData] = React.useState<RecordFormData>({});
  const [errors, setErrors] = React.useState<{ [key: string]: string }>({});

  React.useEffect(() => {
    if (record) {
      setFormData(record.Data);
    } else {
      setFormData({});
    }
    setErrors({});
  }, [record]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const validationErrors = validateRecord(formData, fields);
    if (validationErrors.length > 0) {
      const errorMap: { [key: string]: string } = {};
      validationErrors.forEach(error => {
        errorMap[error.FieldID] = error.Message;
      });
      setErrors(errorMap);
      return;
    }
    onSubmit(formData);
  };

  const handleChange = (field: Field, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field.KeyName]: value
    }));
    if (errors[field.ID]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field.ID];
        return newErrors;
      });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{record ? 'Edit Record' : 'Create Record'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {fields.map((field) => (
            <div key={field.ID}>
              <Label htmlFor={field.ID}>{field.Name}</Label>
              <Input
                id={field.ID}
                value={formData[field.KeyName] || ''}
                onChange={(e) => handleChange(field, e.target.value)}
                required={field.Required}
                type={field.Type === 'number' ? 'number' : 'text'}
              />
              {errors[field.ID] && (
                <p className="text-sm text-red-500 mt-1">{errors[field.ID]}</p>
              )}
            </div>
          ))}
          <div className="flex justify-end space-x-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit">
              {record ? 'Save' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
} 