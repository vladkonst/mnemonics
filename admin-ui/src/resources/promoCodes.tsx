import {
  List, Datagrid, TextField, NumberField, DateField,
  Create, SimpleForm, TextInput, NumberInput, DateTimeInput,
  required, DeleteButton,
} from 'react-admin';

export const PromoCodeList = () => (
  <List sort={{ field: 'created_at', order: 'ASC' }}>
    <Datagrid>
      <TextField source="code" label="Код" />
      <TextField source="university_name" label="Университет" />
      <NumberField source="max_activations" label="Макс. активаций" />
      <NumberField source="remaining" label="Остаток" />
      <TextField source="status" label="Статус" />
      <DateField source="expires_at" label="Истекает" showTime />
      <DateField source="created_at" label="Создан" />
      <DeleteButton label="Деактивировать" />
    </Datagrid>
  </List>
);

const transformPromoCode = (data: any) => ({
  ...data,
  expires_at: data.expires_at ? new Date(data.expires_at).toISOString() : undefined,
});

export const PromoCodeCreate = () => (
  <Create redirect="list" transform={transformPromoCode}>
    <SimpleForm>
      <TextInput source="code" label="Код" validate={required()} fullWidth />
      <TextInput source="university_name" label="Университет" fullWidth />
      <NumberInput source="max_activations" label="Макс. активаций" defaultValue={30} validate={required()} />
      <DateTimeInput source="expires_at" label="Истекает (необязательно)" />
    </SimpleForm>
  </Create>
);
