import type { EdgeInsets } from 'react-native-safe-area-context';

/**
 * Bottom padding for tab scenes so content stays above the floating pill tab bar.
 * Keep in sync with `tabBarStyle` in MainNavigator (bottom offset + height).
 */
export function getFloatingTabBarBottomInset(insets: Pick<EdgeInsets, 'bottom'>): number {
  const tabBarBottom = Math.max(12, insets.bottom + 6);
  const tabBarHeight = 64 + Math.max(0, insets.bottom - 6);
  const gapAboveBar = 10;
  return tabBarBottom + tabBarHeight + gapAboveBar;
}
