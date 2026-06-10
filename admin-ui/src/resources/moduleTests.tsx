import {
  List, Datagrid, NumberField, TextField, DateField,
  DeleteButton, FunctionField,
} from 'react-admin';
import LinkedIdField from '../LinkedIdField';

export const ModuleTestList = () => (
  <List sort={{ field: 'id', order: 'ASC' }} title="Тесты модулей">
    <Datagrid>
      <NumberField source="id" label="ID" />
      <LinkedIdField source="module_id" reference="modules" label="Модуль ID" />
      <TextField source="name" label="Название" />
      <FunctionField label="Вопросов" render={(r: any) => (r.questions || []).length} />
      <NumberField source="passing_score" label="Порог (%)" />
      <DateField source="created_at" label="Создан" />
      <DeleteButton mutationMode="pessimistic" />
    </Datagrid>
  </List>
);
