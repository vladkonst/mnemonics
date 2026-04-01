import { useMemo } from 'react';
import {
  List, Datagrid, NumberField, BooleanField, DateField,
  Create, Edit, SimpleForm, NumberInput, BooleanInput, TextInput,
  SelectInput, ArrayInput, SimpleFormIterator, ReferenceInput,
  required, EditButton, DeleteButton, useGetList,
} from 'react-admin';

const questionTypeChoices = [
  { id: 'multiple_choice', name: 'Несколько вариантов' },
  { id: 'true_false', name: 'Верно/Неверно' },
];

export const TestList = () => (
  <List sort={{ field: 'id', order: 'ASC' }}>
    <Datagrid>
      <NumberField source="id" label="ID" />
      <NumberField source="theme_id" label="Тема ID" />
      <NumberField source="difficulty" label="Сложность" />
      <NumberField source="passing_score" label="Порог (%)" />
      <BooleanField source="shuffle_questions" label="Перемешать вопросы" />
      <BooleanField source="shuffle_answers" label="Перемешать ответы" />
      <DateField source="created_at" label="Создан" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
);

const transformTest = (data: any) => ({
  ...data,
  questions: (data.questions || []).map((q: any, idx: number) => ({
    ...q,
    id: idx + 1,
    order_num: idx + 1,
  })),
});

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

const TestForm = ({ isCreate = false }: { isCreate?: boolean }) => (
  <SimpleForm>
    {isCreate && <ThemeSelectInput />}
    <NumberInput source="difficulty" label="Сложность (1-5)" defaultValue={1} min={1} max={5} validate={required()} />
    <NumberInput source="passing_score" label="Порог (%)" defaultValue={70} min={0} max={100} validate={required()} />
    <BooleanInput source="shuffle_questions" label="Перемешать вопросы" />
    <BooleanInput source="shuffle_answers" label="Перемешать ответы" />
    <ArrayInput source="questions" label="Вопросы">
      <SimpleFormIterator>
        <TextInput source="text" label="Текст вопроса" fullWidth validate={required()} />
        <SelectInput source="type" label="Тип" choices={questionTypeChoices} defaultValue="multiple_choice" />
        <ArrayInput source="options" label="Варианты ответов">
          <SimpleFormIterator>
            <TextInput label="Вариант" source="" />
          </SimpleFormIterator>
        </ArrayInput>
        <TextInput source="correct_answer" label="Правильный ответ" />
      </SimpleFormIterator>
    </ArrayInput>
  </SimpleForm>
);

export const TestCreate = () => (
  <Create redirect="list" transform={transformTest}>
    <TestForm isCreate />
  </Create>
);

export const TestEdit = () => (
  <Edit transform={transformTest}>
    <TestForm />
  </Edit>
);
