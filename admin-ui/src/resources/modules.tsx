import {
  List, Datagrid, TextField, BooleanField, NumberField, DateField,
  Create, Edit, SimpleForm, TextInput, NumberInput, BooleanInput,
  required, EditButton, DeleteButton,
} from 'react-admin';

export const ModuleList = () => (
  <List sort={{ field: 'id', order: 'ASC' }}>
    <Datagrid>
      <NumberField source="id" label="ID" />
      <TextField source="name" label="Название" />
      <TextField source="description" label="Описание" />
      <BooleanField source="is_locked" label="Заблокирован" />
      <TextField source="icon_emoji" label="Иконка" />
      <DateField source="created_at" label="Создан" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export const ModuleCreate = () => (
  <Create redirect="list">
    <SimpleForm>
      <TextInput source="name" label="Название" validate={required()} fullWidth />
      <TextInput source="description" label="Описание" fullWidth multiline />
      <BooleanInput source="is_locked" label="Заблокирован" defaultValue={true} />
      <TextInput source="icon_emoji" label="Иконка (emoji)" />
    </SimpleForm>
  </Create>
);

export const ModuleEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" label="Название" validate={required()} fullWidth />
      <TextInput source="description" label="Описание" fullWidth multiline />
      <NumberInput source="order_num" label="Порядок" />
      <BooleanInput source="is_locked" label="Заблокирован" />
      <TextInput source="icon_emoji" label="Иконка (emoji)" />
    </SimpleForm>
  </Edit>
);
