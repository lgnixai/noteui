import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Calendar } from '@/components/ui/calendar';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { CalendarIcon } from 'lucide-react';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';

import { Field } from '../types/field';
import { Record, RecordFormData } from '../types/record';

interface RecordFormProps {
  tableId: string;
  fields: Field[];
  record?: Record;
  onSubmit: (data: RecordFormData) => Promise<void>;
  onCancel: () => void;
}

export function RecordForm({ tableId, fields, record, onSubmit, onCancel }: RecordFormProps) {
  // Build form schema based on fields
  const formSchema = z.object(
    fields.reduce((acc, field) => {
      let validator: z.ZodType<any>;

      switch (field.type) {
        case 'text':
          validator = z.string();
          if (field.required) {
            validator = validator.min(1, 'This field is required');
          }
          break;
        case 'number':
          validator = z.number({
            invalid_type_error: 'Please enter a valid number',
          });
          if (field.required) {
            validator = validator.min(-Infinity);
          } else {
            validator = validator.nullable();
          }
          break;
        case 'boolean':
          validator = z.boolean();
          if (!field.required) {
            validator = validator.nullable();
          }
          break;
        case 'date':
          validator = z.date({
            invalid_type_error: 'Please enter a valid date',
          });
          if (!field.required) {
            validator = validator.nullable();
          }
          break;
        default:
          validator = z.any();
      }

      return {
        ...acc,
        [field.keyName]: validator,
      };
    }, {})
  );

  type FormData = z.infer<typeof formSchema>;

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: record
      ? fields.reduce((acc, field) => ({
          ...acc,
          [field.keyName]: record.data[field.keyName],
        }), {})
      : fields.reduce((acc, field) => ({
          ...acc,
          [field.keyName]: null,
        }), {}),
  });

  const handleSubmit = async (data: FormData) => {
    await onSubmit(data);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
        {fields.map(field => (
          <FormField
            key={field.id}
            control={form.control}
            name={field.keyName}
            render={({ field: formField }) => (
              <FormItem>
                <FormLabel>{field.name}</FormLabel>
                <FormControl>
                  {field.type === 'text' && (
                    <Input {...formField} />
                  )}
                  {field.type === 'number' && (
                    <Input
                      type="number"
                      {...formField}
                      onChange={e => formField.onChange(e.target.value ? Number(e.target.value) : null)}
                    />
                  )}
                  {field.type === 'boolean' && (
                    <Checkbox
                      checked={formField.value}
                      onCheckedChange={formField.onChange}
                    />
                  )}
                  {field.type === 'date' && (
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button
                          variant="outline"
                          className={cn(
                            'w-full justify-start text-left font-normal',
                            !formField.value && 'text-muted-foreground'
                          )}
                        >
                          <CalendarIcon className="mr-2 h-4 w-4" />
                          {formField.value ? (
                            format(formField.value, 'PPP')
                          ) : (
                            <span>Pick a date</span>
                          )}
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-auto p-0" align="start">
                        <Calendar
                          mode="single"
                          selected={formField.value}
                          onSelect={formField.onChange}
                          initialFocus
                        />
                      </PopoverContent>
                    </Popover>
                  )}
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        ))}

        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            {record ? 'Update' : 'Create'} Record
          </Button>
        </div>
      </form>
    </Form>
  );
} 