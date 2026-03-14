import React, { useState } from 'react';
import { ScrollView, View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, spacing, typography, radii, shadows } from '../theme';

interface AccordionItem {
  title: string;
  icon: keyof typeof Ionicons.glyphMap;
  content: string;
}

const sections: AccordionItem[] = [
  {
    title: 'About the Register',
    icon: 'information-circle',
    content:
      'The HMIS MCH 007 Integrated Family Planning Register (Version 3) is used by all health facilities ' +
      'offering family planning services in Uganda. It captures client demographics, visit type, counseling, ' +
      'contraceptive methods dispensed, side effects, cancer screening, STI screening, and referral details. ' +
      'This mobile app digitizes the register for faster data entry and better data quality.',
  },
  {
    title: 'Client Information (Columns 1–7)',
    icon: 'person',
    content:
      '• Serial Number: Auto-generated per month per facility.\n' +
      '• Client Number: Auto-generated in format FACILITY-DATE-SEQ (e.g., MRRH-260314-001). Visitors do not get a client number.\n' +
      '• NIN: National Identification Number, optional.\n' +
      '• Name: Surname and given name are required.\n' +
      "• Phone: Client's contact number.\n" +
      '• Address: Village, Parish, Subcounty, District.\n' +
      '• Sex: M (Male) or F (Female).\n' +
      "• Age: Client's age in years (0\u2013120).",
  },
  {
    title: 'Visit Type (Columns 8–11)',
    icon: 'clipboard',
    content:
      '• New User: First time receiving FP services. Cannot be selected with Revisit.\n' +
      '• Revisit: Returning for FP services. Must specify the previous method used.\n' +
      '• HTS Code: HIV Testing Services code from the HTS register. Optional.\n\n' +
      'Skip Logic: If "Revisit" is selected, the "Previous Method" field becomes required and visible.',
  },
  {
    title: 'FP Counseling (Columns 12–13)',
    icon: 'chatbubbles',
    content:
      '• Individual: Client was counseled individually.\n' +
      '• As Couple: Client was counseled as part of a couple.\n' +
      '• Counseling Topics: OM (Other Methods), SE (Side Effects), WD (When to Discontinue), MS (Method Switching).\n' +
      '• Switching Method: Check if client is switching from one method to another. Must provide a reason.\n\n' +
      'Skip Logic: "Switching Reason" only appears when "Switching Method" is checked.',
  },
  {
    title: 'Contraceptives Dispensed (Columns 14–20)',
    icon: 'medkit',
    content:
      '• Oral Pills: Record number of cycles for COCs, POPs, and pieces for ECPs.\n' +
      '• Condoms: Record units of male and female condoms.\n' +
      '• Injectables: DMPA-IM, DMPA-SC (Provider Administered), DMPA-SC (Self-Injected) in doses.\n' +
      '• Implants: 3-year and 5-year implant insertions.\n' +
      '• IUDs: Copper-T, Hormonal 3-year, Hormonal 5-year.\n' +
      '• Permanent: Tubal Ligation (females only), Vasectomy (males only).\n' +
      '• FAM: Standard Days Method, LAM, Two Day Method.\n\n' +
      'Skip Logic: Tubal Ligation is hidden for males; Vasectomy is hidden for females.',
  },
  {
    title: 'Post-Pregnancy & LARC Removal (Columns 21–22)',
    icon: 'heart',
    content:
      '• Postpartum FP Timing: When FP was started relative to delivery.\n' +
      '• Post-Abortion FP Timing: When FP was started relative to abortion care.\n' +
      '• Implant/IUD Removal: Record reason and timing if a LARC was removed.\n\n' +
      'Skip Logic: Removal timing only appears when a removal reason is selected.',
  },
  {
    title: 'Side Effects & Screening (Columns 23–25)',
    icon: 'alert-circle',
    content:
      '• Side Effects: Select all codes that apply from the option set.\n' +
      '• Cervical Cancer Screening: Method used, result status, and treatment if positive.\n' +
      '• Breast Cancer Screening: Method and result.\n' +
      '• STI Screening: Whether client was screened for STIs.\n\n' +
      'Skip Logic: Cervical cancer treatment only appears if the status is positive. ' +
      'Cancer screening fields only appear for female clients.',
  },
  {
    title: 'Referral & Remarks (Columns 26–27)',
    icon: 'share',
    content:
      '• Referral Number: If the client was referred, enter the referral slip number.\n' +
      '• Referral Reason: Description of why the client was referred.\n' +
      '• Remarks: Any additional notes for this visit.',
  },
  {
    title: 'User Roles & Data Access',
    icon: 'shield-checkmark',
    content:
      '• Superadmin: Full access to all facilities, users, and audit logs.\n' +
      '• Facility Admin: Manages users and views records for their facility.\n' +
      '• Facility User: Creates and views submissions for their assigned facility.\n' +
      '• Reviewer: Read-only access to submissions at their facility.\n\n' +
      'Data is always scoped by facility. Users can only see data from their assigned facility ' +
      'unless they have superadmin privileges.',
  },
  {
    title: 'Client Number Format',
    icon: 'barcode',
    content:
      'Client numbers are automatically generated with the format:\n' +
      'FACILITY_PREFIX–YYMMDD–SEQ\n\n' +
      'Example: MRRH-260314-001\n\n' +
      '• FACILITY_PREFIX: Abbreviation configured for each facility.\n' +
      '• YYMMDD: Date of visit.\n' +
      '• SEQ: Daily sequence number starting from 001, resets each day.\n' +
      '• Visitors do not receive a client number.',
  },
  {
    title: 'Common Validation Errors',
    icon: 'warning',
    content:
      '• "Must be either new user or revisit": You must check exactly one.\n' +
      '• "Previous method required for revisit": When revisit is checked, select the method.\n' +
      '• "Reason required when switching": If switching is checked, select a reason.\n' +
      '• "Visit date must be YYYY-MM-DD": Use the correct date format.\n' +
      '• "Sex must be M or F": Only valid values are M or F.\n' +
      '• "Age must be between 0 and 120": Enter a valid age.',
  },
];

export function GuideScreen() {
  const [expanded, setExpanded] = useState<number | null>(0);

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>User Guide</Text>
      <Text style={styles.subtitle}>
        HMIS MCH 007 — Integrated Family Planning Register
      </Text>

      {sections.map((section, index) => (
        <View key={index} style={styles.accordion}>
          <TouchableOpacity
            style={styles.accordionHeader}
            onPress={() => setExpanded(expanded === index ? null : index)}
            activeOpacity={0.7}
          >
            <View style={styles.headerLeft}>
              <View style={[styles.iconBox, { backgroundColor: expanded === index ? colors.primary : colors.border }]}>
                <Ionicons name={section.icon} size={16} color={expanded === index ? colors.textInverse : colors.textSecondary} />
              </View>
              <Text style={styles.headerText}>{section.title}</Text>
            </View>
            <Ionicons
              name={expanded === index ? 'chevron-up' : 'chevron-down'}
              size={18}
              color={colors.textMuted}
            />
          </TouchableOpacity>
          {expanded === index && (
            <View style={styles.accordionBody}>
              <Text style={styles.bodyText}>{section.content}</Text>
            </View>
          )}
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  title: { ...typography.h2, color: colors.text, marginBottom: spacing.xs },
  subtitle: { ...typography.bodySmall, color: colors.textMuted, marginBottom: spacing.xxl },
  accordion: {
    backgroundColor: colors.surface,
    borderRadius: radii.md,
    marginBottom: spacing.sm,
    overflow: 'hidden',
    ...shadows.sm,
  },
  accordionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: spacing.lg,
  },
  headerLeft: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, flex: 1 },
  iconBox: {
    width: 28,
    height: 28,
    borderRadius: radii.sm,
    alignItems: 'center',
    justifyContent: 'center',
  },
  headerText: { ...typography.h4, color: colors.text, flex: 1 },
  accordionBody: {
    paddingHorizontal: spacing.lg,
    paddingBottom: spacing.lg,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.divider,
    paddingTop: spacing.md,
  },
  bodyText: { ...typography.body, color: colors.textSecondary, lineHeight: 22 },
});
