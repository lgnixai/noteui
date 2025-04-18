import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useToast } from '@/components/ui/use-toast';

import { Field } from '../types/field';
import { Record, RecordFormData } from '../types/record';
import { recordService } from '../services/record';
import { RecordForm } from './RecordForm';

interface RecordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tableId: string;
  fields: Field[];
  record?: Record;
  onSuccess: () => void;
}

export function RecordDialog({
  open,
  onOpenChange,
  tableId,
  fields,
  record,
  onSuccess,
}: RecordDialogProps) {
  const { toast } = useToast();

  const handleSubmit = async (data: RecordFormData) => {
    try {
      if (record) {
        await recordService.updateRecord(record.id, data);
        toast({
          title: 'Success',
          description: 'Record updated successfully',
        });
      } else {
        await recordService.createRecord(tableId, data);
        toast({
          title: 'Success',
          description: 'Record created successfully',
        });
      }
      onOpenChange(false);
      onSuccess();
    } catch (error) {
      toast({
        title: 'Error',
        description: record
          ? 'Failed to update record'
          : 'Failed to create record',
        variant: 'destructive',
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {record ? 'Edit Record' : 'Create Record'}
          </DialogTitle>
        </DialogHeader>
        <RecordForm
          tableId={tableId}
          fields={fields}
          record={record}
          onSubmit={handleSubmit}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
} 