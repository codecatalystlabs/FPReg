import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Ionicons } from '@expo/vector-icons';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors, typography } from '../theme';
import { useAuthStore } from '../store/authStore';
import { canCreateRegistration } from '../utils/permissions';

import { DashboardScreen } from '../screens/DashboardScreen';
import { SubmissionsScreen } from '../screens/SubmissionsScreen';
import { SubmissionDetailScreen } from '../screens/SubmissionDetailScreen';
import { NewRegistrationScreen } from '../screens/NewRegistrationScreen';
import { EditRegistrationScreen } from '../screens/EditRegistrationScreen';
import { GuideScreen } from '../screens/GuideScreen';
import { ProfileScreen } from '../screens/ProfileScreen';
import { UsersScreen } from '../screens/UsersScreen';
import { CreateUserScreen } from '../screens/CreateUserScreen';

export type MainStackParamList = {
  Tabs: undefined;
  SubmissionDetail: { id: string };
  NewRegistration: undefined;
  EditRegistration: { id: string };
  Users: undefined;
  CreateUser: undefined;
};

export type TabParamList = {
  Home: undefined;
  NewEntry: undefined;
  Submissions: undefined;
  Guide: undefined;
  Profile: undefined;
};

const Stack = createNativeStackNavigator<MainStackParamList>();
const Tab = createBottomTabNavigator<TabParamList>();

function TabNavigator() {
  const user = useAuthStore((s) => s.user);
  const showNewEntry = user ? canCreateRegistration(user.role) : false;
  const insets = useSafeAreaInsets();

  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        tabBarIcon: ({ focused, color, size }) => {
          const icons: Record<string, keyof typeof Ionicons.glyphMap> = {
            Home: focused ? 'home' : 'home-outline',
            NewEntry: focused ? 'add-circle' : 'add-circle-outline',
            Submissions: focused ? 'list' : 'list-outline',
            Guide: focused ? 'book' : 'book-outline',
            Profile: focused ? 'person' : 'person-outline',
          };
          return <Ionicons name={icons[route.name]} size={size} color={color} />;
        },
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textMuted,
        tabBarLabelStyle: { fontSize: 11, fontWeight: '600' },
        tabBarStyle: {
          position: 'absolute',
          left: 16,
          right: 16,
          bottom: Math.max(12, insets.bottom + 6),
          borderRadius: 18,
          borderTopWidth: 0,
          elevation: 8,
          shadowColor: '#000',
          shadowOpacity: 0.08,
          shadowRadius: 8,
          shadowOffset: { width: 0, height: 2 },
          backgroundColor: colors.surface,
          height: 64 + Math.max(0, insets.bottom - 6),
          paddingBottom: 8 + Math.max(0, insets.bottom - 6),
        },
        tabBarSafeAreaInset: { bottom: 0 },
        headerStyle: { backgroundColor: colors.surface },
        headerTitleStyle: { ...typography.h4, color: colors.text },
        headerShadowVisible: false,
      })}
    >
      <Tab.Screen name="Home" component={DashboardScreen} options={{ title: 'Dashboard' }} />
      {showNewEntry && (
        <Tab.Screen name="NewEntry" component={NewRegistrationScreen} options={{ title: 'New Entry' }} />
      )}
      <Tab.Screen name="Submissions" component={SubmissionsScreen} />
      <Tab.Screen name="Guide" component={GuideScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
    </Tab.Navigator>
  );
}

export function MainNavigator() {
  return (
    <Stack.Navigator>
      <Stack.Screen name="Tabs" component={TabNavigator} options={{ headerShown: false }} />
      <Stack.Screen
        name="SubmissionDetail"
        component={SubmissionDetailScreen}
        options={{
          title: 'Submission Detail',
          headerStyle: { backgroundColor: colors.surface },
          headerTitleStyle: { ...typography.h4 },
        }}
      />
      <Stack.Screen
        name="NewRegistration"
        component={NewRegistrationScreen}
        options={{
          title: 'New Registration',
          headerStyle: { backgroundColor: colors.surface },
          headerTitleStyle: { ...typography.h4 },
        }}
      />
      <Stack.Screen
        name="EditRegistration"
        component={EditRegistrationScreen}
        options={{
          title: 'Edit Registration',
          headerStyle: { backgroundColor: colors.surface },
          headerTitleStyle: { ...typography.h4 },
        }}
      />
      <Stack.Screen
        name="Users"
        component={UsersScreen}
        options={{
          title: 'Users',
          headerStyle: { backgroundColor: colors.surface },
          headerTitleStyle: { ...typography.h4 },
        }}
      />
      <Stack.Screen
        name="CreateUser"
        component={CreateUserScreen}
        options={{
          title: 'Create user',
          headerStyle: { backgroundColor: colors.surface },
          headerTitleStyle: { ...typography.h4 },
        }}
      />
    </Stack.Navigator>
  );
}
