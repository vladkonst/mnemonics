import { Admin, Resource } from 'react-admin';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import TopicIcon from '@mui/icons-material/Topic';
import LightbulbIcon from '@mui/icons-material/Lightbulb';
import QuizIcon from '@mui/icons-material/Quiz';
import CardGiftcardIcon from '@mui/icons-material/CardGiftcard';
import PeopleIcon from '@mui/icons-material/People';

import dataProvider from './dataProvider';
import authProvider from './authProvider';
import Dashboard from './Dashboard';
import LoginPage from './LoginPage';

import { ModuleList, ModuleCreate, ModuleEdit } from './resources/modules';
import { ThemeList, ThemeCreate, ThemeEdit } from './resources/themes';
import { MnemonicList, MnemonicCreate, MnemonicEdit } from './resources/mnemonics';
import { TestList, TestCreate, TestEdit } from './resources/tests';
import { PromoCodeList, PromoCodeCreate } from './resources/promoCodes';
import { UserList, UserCreate, UserEdit } from './resources/users';

const App = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    dashboard={Dashboard}
    loginPage={LoginPage}
    title="Mnemo Admin"
  >
    <Resource
      name="modules"
      list={ModuleList}
      create={ModuleCreate}
      edit={ModuleEdit}
      icon={ViewModuleIcon}
      options={{ label: 'Модули' }}
    />
    <Resource
      name="themes"
      list={ThemeList}
      create={ThemeCreate}
      edit={ThemeEdit}
      icon={TopicIcon}
      options={{ label: 'Темы' }}
    />
    <Resource
      name="mnemonics"
      list={MnemonicList}
      create={MnemonicCreate}
      edit={MnemonicEdit}
      icon={LightbulbIcon}
      options={{ label: 'Мнемоники' }}
    />
    <Resource
      name="tests"
      list={TestList}
      create={TestCreate}
      edit={TestEdit}
      icon={QuizIcon}
      options={{ label: 'Тесты' }}
    />
    <Resource
      name="promo_codes"
      list={PromoCodeList}
      create={PromoCodeCreate}
      icon={CardGiftcardIcon}
      options={{ label: 'Промокоды' }}
    />
    <Resource
      name="users"
      list={UserList}
      create={UserCreate}
      edit={UserEdit}
      icon={PeopleIcon}
      options={{ label: 'Пользователи' }}
    />
  </Admin>
);

export default App;
