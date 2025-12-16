"use client";

import React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ZodObject } from "zod";
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
  FormDescription,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import MultipleSelector from "@/components/ui/multi-select";
import { Slider } from "@/components/ui/slider";

// Helper type to extract schema keys
type SchemaKeys<T extends ZodObject<any>> = keyof T["shape"];

type FieldConfig<T extends ZodObject<any> = ZodObject<any>> = {
  name: SchemaKeys<T>;
  label?: string;
  placeholder?: string;
  type?:
    | "text"
    | "email"
    | "password"
    | "textarea"
    | "date"
    | "select"
    | "multi-select"
    | "number"
    | "slider";
  options?: string[];
  description?: string;
  min?: number;
  max?: number;
  step?: number;
};

// Utility function to format field names (e.g., rolesNoEnum -> Roles No Enum)
const formatFieldName = (fieldName: string): string => {
  return fieldName
    .replace(/([A-Z])/g, " $1") // Add space before uppercase
    .replace(/^./, (str) => str.toUpperCase()) // Capitalize first letter
    .trim();
};

// interface FormGeneratorReturn<T extends ZodObject<any> = ZodObject<any>> {
//   form: React.FC<React.HTMLAttributes<HTMLFormElement>>
//   field: React.FC<FieldConfig<T> & React.HTMLAttributes<HTMLDivElement>>
//   errors: React.FC
//   submit: React.FC<React.ComponentProps<typeof Button>>
//   reset: React.FC<React.ComponentProps<typeof Button>>
//   close: React.FC<React.ComponentProps<typeof Button>>
//   FullForm: React.FC
// }

export type FormGeneratorInputs<T extends ZodObject<any>> = {
  schema: T;
  formConfig: {
    formTitle?: string;
    formDescription?: string;
    onSubmit: (data: T["_output"]) => void;
    defaultValues?: Partial<T["_output"]>;
    onReset?: () => void;
  };
};

export function useFormGenerator<T extends ZodObject<any>>(
  schema: T,
  formConfig: {
    formTitle?: string;
    formDescription?: string;
    onSubmit: (data: T["_output"]) => void;
    defaultValues?: Partial<T["_output"]>;
    onReset?: () => void;
  },
) {
  // Check if a field is optional
  const isFieldOptional = (fieldName: string): boolean => {
    const fieldSchema = schema.shape[fieldName];
    if (!fieldSchema) return false;
    return fieldSchema.isOptional?.() === true;
  };

  // Clean up empty strings for optional fields before submission
  const cleanDataBeforeSubmit = (data: any): T["_output"] => {
    const cleaned = { ...data };
    Object.keys(cleaned).forEach((key) => {
      const fieldSchema = schema.shape[key];

      // Convert empty strings to undefined for optional fields
      if (isFieldOptional(key) && cleaned[key] === "") {
        cleaned[key] = undefined;
      }

      // Handle empty date strings for optional date fields
      if (
        isFieldOptional(key) &&
        fieldSchema.def?.type === "date" &&
        cleaned[key] === ""
      ) {
        cleaned[key] = undefined;
      }

      // Handle empty number strings for optional number fields
      if (
        isFieldOptional(key) &&
        fieldSchema.def?.type === "number" &&
        cleaned[key] === ""
      ) {
        cleaned[key] = undefined;
      }
    });
    return cleaned;
  };

  const form = useForm({
    resolver: zodResolver(schema),
    mode: "onBlur",
    defaultValues: Object.keys(schema.shape).reduce((acc, key) => {
      const fieldSchema = schema.shape[key];

      // Use provided default value if available
      if (formConfig.defaultValues && key in formConfig.defaultValues) {
        acc[key] = formConfig.defaultValues[key as keyof T["_output"]];
        return acc;
      }

      if (fieldSchema.def?.type === "array") {
        acc[key] = [];
      } else if (fieldSchema.def?.type === "date") {
        // For date fields, initialize as empty string
        acc[key] = "";
      } else if (fieldSchema.def?.type === "number") {
        // For number fields, use empty string initially
        acc[key] = "";
      } else {
        acc[key] = "";
      }
      return acc;
    }, {} as any),
  });

  // Extract field information from schema
  const getFieldType = (fieldName: string): string => {
    const fieldSchema = schema.shape[fieldName];
    if (!fieldSchema) return "text";

    // Unwrap optional types
    let actualSchema = fieldSchema;
    if (fieldSchema.isOptional?.() === true) {
      actualSchema = fieldSchema.unwrap?.() || fieldSchema;
    }

    // Check if it's directly an array type
    if (actualSchema.def?.type === "array") {
      const arrayElement = actualSchema.element;
      // Check if array contains enums
      if (arrayElement?.def?.type === "enum") {
        return "multi-select";
      }
      // Array of strings
      return "multi-select";
    }

    const typeName = actualSchema.def?.type;

    if (typeName === "enum") {
      return "select";
    }
    if (typeName === "date") {
      return "date";
    }
    if (typeName === "number") {
      return "number";
    }
    if (typeName === "string") {
      const checks = actualSchema.def?.checks || [];
      for (const check of checks) {
        if (check.kind === "email") return "email";
      }
      // Infer password field from field name
      if (
        fieldName.toLowerCase().includes("password") ||
        fieldName.toLowerCase().includes("pwd")
      ) {
        return "password";
      }
      return "text";
    }

    return "text";
  };

  const getFieldEnumOptions = (fieldName: string): string[] => {
    const fieldSchema = schema.shape[fieldName];
    if (!fieldSchema) return [];

    // Direct array type
    if (fieldSchema.def?.type === "array") {
      const arrayElement = fieldSchema.element;
      if (arrayElement?.def?.type === "enum") {
        return arrayElement.options || [];
      }
      return [];
    }

    const typeName = fieldSchema.def?.type;

    if (typeName === "enum") {
      return fieldSchema.options || [];
    }

    return [];
  };

  // Render individual field inputs
  const renderFieldInput = (
    fieldName: string,
    fieldConfig: FieldConfig,
    field: any,
  ) => {
    const defaultType = getFieldType(fieldName);
    const type = fieldConfig.type || defaultType;
    const options = fieldConfig.options || getFieldEnumOptions(fieldName);

    switch (type) {
      case "textarea":
        return (
          <Textarea
            placeholder={fieldConfig.placeholder || `Enter ${fieldName}...`}
            value={field.value || ""}
            onChange={field.onChange}
            onBlur={field.onBlur}
          />
        );
      case "password":
        return (
          <Input
            type="password"
            placeholder={fieldConfig.placeholder || "••••••••"}
            value={field.value || ""}
            onChange={field.onChange}
            onBlur={field.onBlur}
          />
        );
      case "email":
        return (
          <Input
            type="email"
            placeholder={fieldConfig.placeholder || "john@example.com"}
            value={field.value || ""}
            onChange={field.onChange}
            onBlur={field.onBlur}
          />
        );
      case "date":
        return (
          <Input
            type="date"
            value={
              field.value instanceof Date
                ? field.value.toISOString().split("T")[0]
                : typeof field.value === "string"
                  ? field.value
                  : ""
            }
            onChange={(e) => {
              const dateString = e.target.value;
              if (dateString) {
                const date = new Date(dateString + "T00:00:00");
                field.onChange(date);
              } else {
                // Keep as empty string, will be converted to undefined on submit if optional
                field.onChange("");
              }
            }}
            onBlur={field.onBlur}
          />
        );
      case "number":
        return (
          <Input
            type="number"
            placeholder={fieldConfig.placeholder || "Enter a number..."}
            value={field.value ?? ""}
            onChange={(e) => {
              const value = e.target.value;
              if (value === "") {
                field.onChange(isFieldOptional(fieldName) ? undefined : "");
              } else {
                field.onChange(Number(value));
              }
            }}
            onBlur={field.onBlur}
          />
        );
      case "slider":
        const min = fieldConfig.min ?? 0;
        const max = fieldConfig.max ?? 100;
        const step = fieldConfig.step ?? 1;

        return (
          <div className="flex items-center gap-4">
            <Slider
              min={min}
              max={max}
              step={step}
              value={[field.value ?? min]}
              onValueChange={(value) => {
                field.onChange(value[0]);
              }}
              className="flex-1"
            />
            <span className="text-sm font-medium min-w-12 text-right">
              {field.value ?? "-"}
            </span>
          </div>
        );
      case "multi-select":
        return (
          <MultipleSelector
            value={
              Array.isArray(field.value)
                ? field.value.map((v: string) => ({ value: v, label: v }))
                : []
            }
            options={options.map((opt) => ({ value: opt, label: opt }))}
            onChange={(selectedOptions) => {
              field.onChange(selectedOptions.map((opt) => opt.value));
            }}
            placeholder={fieldConfig.placeholder || `Select ${fieldName}...`}
            creatable={options.length === 0}
            emptyIndicator={
              options.length === 0 ? (
                <p className="w-full text-center text-sm leading-10 text-muted-foreground">
                  no results found. type to create new.
                </p>
              ) : undefined
            }
          />
        );
      case "select":
        return (
          <Select value={field.value || ""} onValueChange={field.onChange}>
            <SelectTrigger>
              <SelectValue
                placeholder={
                  fieldConfig.placeholder || `Select ${fieldName}...`
                }
              />
            </SelectTrigger>
            <SelectContent>
              {options.map((option: string) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        );
      case "text":
      default:
        return (
          <Input
            type="text"
            placeholder={fieldConfig.placeholder || `Enter ${fieldName}...`}
            value={field.value || ""}
            onChange={field.onChange}
            onBlur={field.onBlur}
          />
        );
    }
  };

  const FormComponentImpl = React.forwardRef<
    HTMLFormElement,
    React.HTMLAttributes<HTMLFormElement>
  >(({ className, children, ...props }, ref) => {
    const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();

      // Get raw form values
      const rawData = form.getValues();

      // Clean up empty optional fields before validation
      const cleanedData = cleanDataBeforeSubmit(rawData);

      // Validate the cleaned data
      try {
        const validatedData = await schema.parseAsync(cleanedData);
        formConfig.onSubmit(validatedData);
      } catch (error) {
        // Set form errors from validation
        if (error instanceof Error && "issues" in error) {
          const issues = (error as any).issues;
          issues.forEach((issue: any) => {
            form.setError(issue.path[0], {
              message: issue.message,
            });
          });
        }
      }
    };

    return (
      <Form {...form}>
        <form
          ref={ref}
          onSubmit={handleFormSubmit}
          className={className}
          {...props}
        >
          {children}
        </form>
      </Form>
    );
  });

  FormComponentImpl.displayName = "FormComponent";

  const ErrorsComponent = () => {
    const errors = form.formState.errors;
    const errorCount = Object.keys(errors).length;

    if (errorCount === 0) return null;

    return (
      <div className="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
        <p className="font-medium">{errorCount} error(s) found</p>
      </div>
    );
  };

  const FieldComponentImpl = React.forwardRef<
    HTMLDivElement,
    FieldConfig<T> & React.HTMLAttributes<HTMLDivElement>
  >(
    (
      {
        name,
        label,
        placeholder,
        type,
        options,
        description,
        min,
        max,
        step,
        ...props
      },
      ref,
    ) => {
      const nameStr = String(name);
      const defaultType = getFieldType(nameStr) as
        | "text"
        | "email"
        | "password"
        | "textarea"
        | "date"
        | "select"
        | "multi-select"
        | "number"
        | "slider";
      const finalType = (type || defaultType) as
        | "text"
        | "email"
        | "password"
        | "textarea"
        | "date"
        | "select"
        | "multi-select"
        | "number"
        | "slider";
      const finalLabel = label || formatFieldName(nameStr);

      return (
        <FormField
          control={form.control}
          name={nameStr}
          render={({ field }) => (
            <FormItem ref={ref} {...props}>
              <FormLabel>{finalLabel}</FormLabel>
              <FormControl>
                {renderFieldInput(
                  nameStr,
                  {
                    name,
                    label: finalLabel,
                    placeholder,
                    type: finalType,
                    options,
                    min,
                    max,
                    step,
                  },
                  field,
                )}
              </FormControl>
              {description && <FormDescription>{description}</FormDescription>}
              <FormMessage />
            </FormItem>
          )}
        />
      );
    },
  );

  FieldComponentImpl.displayName = "FieldComponent";

  const SubmitComponentImpl = React.forwardRef<
    HTMLButtonElement,
    React.ComponentProps<typeof Button>
  >((props, ref) => <Button ref={ref} type="submit" {...props} />);

  SubmitComponentImpl.displayName = "SubmitComponent";

  const ResetComponentImpl = React.forwardRef<
    HTMLButtonElement,
    React.ComponentProps<typeof Button>
  >((props, ref) => (
    <Button
      ref={ref}
      type="button"
      variant="outline"
      onClick={() => {
        form.reset();
        formConfig.onReset?.();
      }}
      {...props}
    />
  ));

  ResetComponentImpl.displayName = "ResetComponent";

  const fields = Object.keys(schema.shape) as (SchemaKeys<T> & string)[];
  const FullFormComponent = () => {
    return (
      <FormComponentImpl className="space-y-4">
        {formConfig.formTitle && (
          <h2 className="text-lg font-semibold">{formConfig.formTitle}</h2>
        )}
        {formConfig.formDescription && (
          <p className="text-sm text-muted-foreground">
            {formConfig.formDescription}
          </p>
        )}
        <ErrorsComponent />
        {fields.map((fieldName) => (
          <FieldComponentImpl key={fieldName} name={fieldName} />
        ))}
        <div className="flex gap-2 pt-4">
          <SubmitComponentImpl>Submit</SubmitComponentImpl>
          <ResetComponentImpl>Reset</ResetComponentImpl>
        </div>
      </FormComponentImpl>
    );
  };
  const fragmentedForm = {
    fieldsList: fields,
    fields: (
      <>
        {fields.map((fieldName) => (
          <FieldComponentImpl key={fieldName} name={fieldName} />
        ))}
      </>
    ),
    errors: () => <ErrorsComponent />,
    formTitle: formConfig.formTitle ? (
      <h2 className="text-lg font-semibold">{formConfig.formTitle}</h2>
    ) : null,
    formDescription: formConfig.formDescription ? (
      <p className="text-sm text-muted-foreground">
        {formConfig.formDescription}
      </p>
    ) : null,
    wrapper: (
      props?: React.HTMLAttributes<HTMLFormElement> & {
        children: React.ReactNode;
      },
    ) => (
      <FormComponentImpl className="space-y-4" {...props}>
        {props?.children}
      </FormComponentImpl>
    ),
    field: (
      fieldProps: FieldConfig<T> & React.HTMLAttributes<HTMLDivElement>,
    ) => <FieldComponentImpl {...fieldProps} />,

    // Also expose prop-based versions for customization
    submitButton: (props?: React.ComponentProps<typeof Button>) => (
      <SubmitComponentImpl {...props}>Submit</SubmitComponentImpl>
    ),
    resetButton: (props?: React.ComponentProps<typeof Button>) => (
      <ResetComponentImpl {...props}>Reset</ResetComponentImpl>
    ),
  };

  return {
    wrapper: FormComponentImpl,
    form: form,
    field: FieldComponentImpl,
    errors: ErrorsComponent,
    submit: SubmitComponentImpl,
    reset: ResetComponentImpl,
    FullForm: FullFormComponent,
    parts: fragmentedForm,
  };
}
