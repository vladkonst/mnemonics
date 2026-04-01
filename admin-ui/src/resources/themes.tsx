import {
  List, Datagrid, TextField, BooleanField, NumberField, DateField,
  Create, Edit, SimpleForm, TextInput, NumberInput, BooleanInput,
  ReferenceInput, SelectInput,
  required, EditButton, DeleteButton,
} from 'react-admin';

export const ThemeList = () => (
  <List sort={{ field: 'id', order: 'ASC' }}>
    <Datagrid>
      <NumberField source="id" label="ID" />
      <NumberField source="module_id" label="Модуль ID" />
      <TextField source="name" label="Название" />
      <BooleanField source="is_introduction" label="Введение" />
      <BooleanField source="is_locked" label="Заблокирована" />
      <NumberField source="estimated_time_minutes" label="Время (мин)" />
      <DateField source="created_at" label="Создана" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

export const ThemeCreate = () => (
  <Create redirect="list">
    <SimpleForm>
      <ReferenceInput source="module_id" reference="modules">
        <SelectInput optionText="name" label="Модуль" validate={required()} />
      </ReferenceInput>
      <TextInput source="name" label="Название" validate={required()} fullWidth />
      <TextInput source="description" label="Описание" fullWidth multiline />
      <BooleanInput source="is_introduction" label="Является введением" />
      <BooleanInput source="is_locked" label="Заблокирована" defaultValue={true} />
      <NumberInput source="estimated_time_minutes" label="Оценочное время (мин)" />
    </SimpleForm>
  </Create>
);

export const ThemeEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="name" label="Название" validate={required()} fullWidth />
      <TextInput source="description" label="Описание" fullWidth multiline />
      <NumberInput source="order_num" label="Порядок" />
      <BooleanInput source="is_introduction" label="Является введением" />
      <BooleanInput source="is_locked" label="Заблокирована" />
      <NumberInput source="estimated_time_minutes" label="Оценочное время (мин)" />
    </SimpleForm>
  </Edit>
);
