import React, { useState, useRef, createContext, useContext, useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import {
  List, Datagrid, TextField, NumberField, DateField,
  Create, Edit, SimpleForm, TextInput, NumberInput, SelectInput,
  ReferenceInput,
  required, EditButton, DeleteButton,
  useInput, InputProps, useRecordContext, useGetList,
} from 'react-admin';

const typeChoices = [
  { id: 'text', name: 'Текст' },
  { id: 'image', name: 'Изображение' },
];

// ── Pending file context (deferred upload) ─────────────────────────────────────

const PendingFileContext = createContext<React.MutableRefObject<File | null> | null>(null);

const uploadToS3 = async (file: File): Promise<string> => {
  const token = localStorage.getItem('admin_token') || '';
  const formData = new FormData();
  formData.append('file', file);
  const res = await fetch('/api/v1/admin/upload', {
    method: 'POST',
    headers: { 'X-Admin-Token': token },
    body: formData,
  });
  const data = await res.json();
  return data.key;
};

// ── Image upload input (stores file, uploads on save) ─────────────────────────

const ImageUploadInput = (props: InputProps & { label?: string }) => {
  const { field } = useInput(props);
  const [localPreview, setLocalPreview] = useState<string | null>(null);
  const pendingFileRef = useContext(PendingFileContext);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (pendingFileRef) pendingFileRef.current = file;
    setLocalPreview(URL.createObjectURL(file));
    field.onChange('__pending__');
  };

  const existingKey = field.value && field.value !== '__pending__' ? field.value : null;

  return (
    <div style={{ marginBottom: 16 }}>
      <div style={{ fontSize: 12, color: isDark ? 'rgba(255,255,255,0.7)' : 'rgba(0,0,0,0.6)', marginBottom: 4 }}>
        {props.label ?? 'S3 изображение'}
      </div>
      <input type="file" accept="image/*" onChange={handleFileChange} />
      {localPreview && (
        <img src={localPreview} alt="preview" style={{ maxHeight: 120, maxWidth: 200, display: 'block', marginTop: 8 }} />
      )}
      {!localPreview && existingKey && (
        <div style={{ marginTop: 8 }}>
          <img
            src={`/api/v1/uploads/${existingKey}`}
            alt="preview"
            style={{ maxHeight: 120, maxWidth: 200, display: 'block', marginBottom: 4 }}
          />
          <span style={{ fontSize: 11, color: isDark ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.4)' }}>{existingKey}</span>
        </div>
      )}
    </div>
  );
};

// ── S3 image as clickable link in list ────────────────────────────────────────

const S3ImageLinkField = ({ source }: { source: string }) => {
  const record = useRecordContext();
  if (!record?.[source]) return <span style={{ color: 'rgba(0,0,0,0.3)' }}>—</span>;
  const key = record[source];
  return (
    <a href={`/api/v1/uploads/${key}`} target="_blank" rel="noopener noreferrer" style={{ fontSize: 13 }}>
      {key}
    </a>
  );
};
S3ImageLinkField.defaultProps = { label: 'S3 ключ' };

// ── Theme selector with module name ───────────────────────────────────────────

const ThemeSelectInput = () => {
  const { data: modules } = useGetList('modules', {
    pagination: { page: 1, perPage: 1000 },
    sort: { field: 'id', order: 'ASC' },
  });
  const moduleMap = useMemo(() => {
    const map: Record<number, string> = {};
    (modules || []).forEach((m: any) => { map[m.id] = m.name; });
    return map;
  }, [modules]);

  return (
    <ReferenceInput source="theme_id" reference="themes">
      <SelectInput
        optionText={(r: any) => `${r.name} (${moduleMap[r.module_id] ?? `модуль ${r.module_id}`})`}
        label="Тема"
        validate={required()}
      />
    </ReferenceInput>
  );
};

// ── Transform: upload pending file before save ────────────────────────────────

const makeTransform = (pendingFileRef: React.MutableRefObject<File | null>) =>
  async (data: any) => {
    if (pendingFileRef.current) {
      const key = await uploadToS3(pendingFileRef.current);
      pendingFileRef.current = null;
      return { ...data, s3_image_key: key };
    }
    if (data.s3_image_key === '__pending__') {
      const { s3_image_key, ...rest } = data;
      return rest;
    }
    return data;
  };

// ── Resources ─────────────────────────────────────────────────────────────────

export const MnemonicList = () => (
  <List sort={{ field: 'id', order: 'ASC' }}>
    <Datagrid>
      <NumberField source="id" label="ID" />
      <NumberField source="theme_id" label="Тема ID" />
      <TextField source="type" label="Тип" />
      <TextField source="content_text" label="Текст" />
      <S3ImageLinkField source="s3_image_key" />
      <DateField source="created_at" label="Создана" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export const MnemonicCreate = () => {
  const pendingFileRef = useRef<File | null>(null);
  return (
    <PendingFileContext.Provider value={pendingFileRef}>
      <Create redirect="list" transform={makeTransform(pendingFileRef)}>
        <SimpleForm>
          <ThemeSelectInput />
          <SelectInput source="type" label="Тип" choices={typeChoices} validate={required()} />
          <TextInput source="content_text" label="Текст" fullWidth multiline />
          <ImageUploadInput source="s3_image_key" label="Изображение (загрузить в S3)" />
        </SimpleForm>
      </Create>
    </PendingFileContext.Provider>
  );
};

export const MnemonicEdit = () => {
  const pendingFileRef = useRef<File | null>(null);
  return (
    <PendingFileContext.Provider value={pendingFileRef}>
      <Edit transform={makeTransform(pendingFileRef)}>
        <SimpleForm>
          <TextInput source="content_text" label="Текст" fullWidth multiline />
          <ImageUploadInput source="s3_image_key" label="Изображение (загрузить в S3)" />
          <NumberInput source="order_num" label="Порядок" />
        </SimpleForm>
      </Edit>
    </PendingFileContext.Provider>
  );
};
