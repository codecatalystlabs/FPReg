import type { NavigationProp, ParamListBase } from '@react-navigation/native';
import type { MainStackParamList } from './MainNavigator';

/** Tab (and other nested) screens must use this to open routes on the root stack. */
export function navigateToMainStack<RouteName extends keyof MainStackParamList>(
  navigation: NavigationProp<ParamListBase>,
  screen: RouteName,
  params?: MainStackParamList[RouteName],
): void {
  const root = navigation.getParent() ?? navigation;
  root.navigate(screen as string, params as object | undefined);
}
