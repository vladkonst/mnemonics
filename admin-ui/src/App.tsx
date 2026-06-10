import { Admin, Resource } from 'react-admin';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import TopicIcon from '@mui/icons-material/Topic';
import LightbulbIcon from '@mui/icons-material/Lightbulb';
import QuizIcon from '@mui/icons-material/Quiz';
import LinkIcon from '@mui/icons-material/Link';
import PeopleIcon from '@mui/icons-material/People';
import FeedbackIcon from '@mui/icons-material/Feedback';
import BusinessIcon from '@mui/icons-material/Business';
import GroupsIcon from '@mui/icons-material/Groups';

import dataProvider from './dataProvider';
import authProvider from './authProvider';
import i18nProvider from './i18nProvider';
import Dashboard from './Dashboard';
import LoginPage from './LoginPage';

import { ModuleList, ModuleCreate, ModuleEdit } from './resources/modules';
import { ThemeList, ThemeCreate, ThemeEdit } from './resources/themes';
import { MnemonicList, MnemonicCreate, MnemonicEdit } from './resources/mnemonics';
import { TestList, TestCreate, TestEdit } from './resources/tests';
import { UserList, UserCreate, UserEdit, UserShow } from './resources/users';
import { FeedbackList, FeedbackShow } from './resources/feedback';
import { InviteLinkList, InviteLinkShow } from './resources/inviteLinks';
import { CorporatePurchaseList, CorporatePurchaseShow } from './resources/corporatePurchases';
import { CorporateGroupList, CorporateGroupShow } from './resources/corporateGroups';

const App = () => (
  <Admin
    dataProvider={dataProvider}
    authProvider={authProvider}
    i18nProvider={i18nProvider}
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
      options={{ label: 'Тесты тем' }}
    />
    <Resource
      name="users"
      list={UserList}
      show={UserShow}
      create={UserCreate}
      edit={UserEdit}
      icon={PeopleIcon}
      options={{ label: 'Пользователи' }}
    />
    <Resource
      name="feedback"
      list={FeedbackList}
      show={FeedbackShow}
      icon={FeedbackIcon}
      options={{ label: 'Обратная связь' }}
    />
    <Resource
      name="invite_links"
      list={InviteLinkList}
      show={InviteLinkShow}
      icon={LinkIcon}
      options={{ label: 'Ссылки-приглашения' }}
    />
    <Resource
      name="corporate_purchases"
      list={CorporatePurchaseList}
      show={CorporatePurchaseShow}
      icon={BusinessIcon}
      options={{ label: 'Корп. покупки' }}
    />
    <Resource
      name="corporate_groups"
      list={CorporateGroupList}
      show={CorporateGroupShow}
      icon={GroupsIcon}
      options={{ label: 'Корп. группы' }}
    />
  </Admin>
);

export default App;
