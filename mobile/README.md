# FP Register Mobile App

React Native (Expo) mobile application for the HMIS MCH 007 Integrated Family Planning Register system.

## Tech Stack

- **React Native** with Expo SDK 52
- **TypeScript** for type safety
- **React Navigation** (native stack + bottom tabs)
- **Zustand** for state management
- **Axios** with token interceptor for API communication
- **React Hook Form** + **Zod** for form validation
- **expo-secure-store** for secure JWT token persistence

## Project Structure

```
mobile/
├── App.tsx                    # Entry point
├── src/
│   ├── api/                   # API service layer
│   │   ├── client.ts          # Axios instance with auth interceptor
│   │   ├── auth.ts            # Auth endpoints
│   │   ├── registrations.ts   # Registration CRUD
│   │   ├── optionSets.ts      # Option set lookup
│   │   └── facilities.ts      # Facility endpoints
│   ├── components/            # Reusable UI components
│   │   ├── AppButton.tsx
│   │   ├── AppInput.tsx
│   │   ├── AppSelect.tsx
│   │   ├── AppCheckbox.tsx
│   │   ├── AppCard.tsx
│   │   ├── SectionHeader.tsx
│   │   ├── StatusBadge.tsx
│   │   ├── EmptyState.tsx
│   │   ├── LoadingState.tsx
│   │   ├── ErrorState.tsx
│   │   ├── SubmissionListItem.tsx
│   │   └── PermissionGuard.tsx
│   ├── navigation/
│   │   ├── AppNavigator.tsx   # Root navigator (auth check)
│   │   ├── AuthNavigator.tsx  # Login stack
│   │   └── MainNavigator.tsx  # Tabs + detail stacks
│   ├── screens/
│   │   ├── SplashScreen.tsx
│   │   ├── LoginScreen.tsx
│   │   ├── DashboardScreen.tsx
│   │   ├── SubmissionsScreen.tsx
│   │   ├── SubmissionDetailScreen.tsx
│   │   ├── NewRegistrationScreen.tsx
│   │   ├── EditRegistrationScreen.tsx
│   │   ├── ProfileScreen.tsx
│   │   └── GuideScreen.tsx
│   ├── store/
│   │   ├── authStore.ts       # Auth state + secure storage
│   │   └── optionSetStore.ts  # Option set cache
│   ├── theme/index.ts         # Colors, spacing, typography, shadows
│   ├── types/index.ts         # TypeScript interfaces
│   ├── constants/index.ts     # API URL, storage keys, roles
│   └── utils/
│       ├── permissions.ts     # Role-based permission helpers
│       ├── logger.ts          # Structured logging
│       └── format.ts          # Date/text formatters
```

## Quick Start

### Prerequisites

- Node.js 18+
- Expo CLI (`npm install -g expo-cli`)
- Backend API running (see main project README)

### Setup

```bash
cd mobile
npm install
```

### Configure API URL

Edit `src/constants/index.ts` and set the API base URL to point to your backend:

- **Android emulator**: `http://10.0.2.2:8080/api/v1`
- **iOS simulator**: `http://localhost:8080/api/v1`
- **Physical device**: `http://<your-machine-ip>:8080/api/v1`

### Run

```bash
npx expo start
```

Then press `a` for Android or `i` for iOS.

## Screens

| Screen | Description |
|--------|-------------|
| Splash | Session restoration with branded loading |
| Login | Email + password with JWT auth |
| Dashboard | Today's stats, recent entries, quick actions |
| New Registration | Full 27-column data entry form with skip logic |
| Submissions | Searchable, paginated list with pull-to-refresh |
| Submission Detail | Read-only sectioned view with edit/delete |
| Edit Registration | Pre-filled form for updating records |
| Profile | User info, facility, role, sign out |
| Guide | Accordion-style user manual |

## Role-Based Access

| Feature | superadmin | facility_admin | facility_user | reviewer |
|---------|:----------:|:--------------:|:-------------:|:--------:|
| Dashboard | Yes | Yes | Yes | Yes |
| Create Registration | Yes | Yes | Yes | No |
| Edit Registration | Yes | Yes | Yes | No |
| Delete Registration | Yes | Yes | No | No |
| View Submissions | All | Facility | Facility | Facility |
| Manage Users | Yes | Yes | No | No |

## Authentication Flow

1. User enters email + password on Login screen
2. App calls `POST /api/v1/auth/login`
3. Access token + refresh token stored in `expo-secure-store`
4. Axios interceptor attaches bearer token to all requests
5. On 401, interceptor automatically attempts token refresh
6. Failed refresh clears session and redirects to Login

## Skip Logic (Form)

- **Previous Method**: shown only when Revisit is checked
- **Switching Reason**: shown only when Switching is checked
- **Tubal Ligation**: hidden for male clients
- **Vasectomy**: hidden for female clients
- **Cervical/Breast Screening**: shown only for female clients
- **Cervical Treatment**: shown only when cervical status is positive
- **LARC Removal Timing**: shown only when a removal reason is selected

## Offline Readiness

The architecture supports future offline capability:

- API calls are isolated in `src/api/` services
- Option sets are cached in `AsyncStorage` for fallback
- State management is separate from API layer
- Form data can be stored locally as drafts (future)

To add offline sync:
1. Queue submissions in local storage when offline
2. Sync queue on connectivity restore
3. Add conflict resolution for concurrent edits

## Security

- Tokens stored in `expo-secure-store` (encrypted at rest)
- No secrets hardcoded in source
- API client handles token refresh transparently
- Protected screens require authentication
- Role-based UI filtering via `PermissionGuard` component
- Backend remains the source of truth for all authorization
