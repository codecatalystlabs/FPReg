import type { MainStackParamList } from './MainNavigator';

/** Minimal nav shape so nested tab screens can call `getParent()` without RootParamList clashes. */
type NavChain = {
  getParent?: () => NavChain | undefined;
  navigate: (name: string, params?: object | undefined) => void;
};

/** Tab (and other nested) screens must use this to open routes on the root stack. */
export function navigateToMainStack<RouteName extends keyof MainStackParamList>(
  navigation: NavChain,
  screen: RouteName,
  params?: MainStackParamList[RouteName],
): void {
  const root = navigation.getParent?.() ?? navigation;
  root.navigate(screen as string, params as object | undefined);
}
