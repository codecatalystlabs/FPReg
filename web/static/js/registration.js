document.addEventListener('DOMContentLoaded', async () => {
  if (!App.requireAuth()) return;
  App.initSidebar();

  let optionSets = {};
  try {
    const data = await App.api('GET', '/option-sets');
    optionSets = data.data;
  } catch (e) {
    App.showToast('Failed to load option sets', 'error');
  }

  populateSelects(optionSets);
  setupSkipLogic();
  setupFormSubmission();
  setDefaults();
});

function setDefaults() {
  const today = new Date().toISOString().split('T')[0];
  const visitDate = document.getElementById('visit_date');
  if (visitDate && !visitDate.value) visitDate.value = today;
}

function populateSelects(opts) {
  populateSelect('hts_code', opts.hts_code);
  populateSelect('previous_method', opts.fp_method);
  populateSelect('switching_reason', opts.switching_reason);
  populateSelect('postpartum_fp_timing', opts.postpartum_fp_timing);
  populateSelect('post_abortion_fp_timing', opts.post_abortion_fp_timing);
  populateSelect('implant_removal_reason', opts.larc_removal_reason);
  populateSelect('implant_removal_timing', opts.larc_removal_timing);
  populateSelect('iud_removal_reason', opts.larc_removal_reason);
  populateSelect('iud_removal_timing', opts.larc_removal_timing);
  populateSelect('cervical_screening_method', opts.cervical_screening_method);
  populateSelect('cervical_cancer_status', opts.cervical_cancer_status);
  populateSelect('cervical_cancer_treatment', opts.cervical_cancer_treatment);
  populateSelect('breast_cancer_screening', opts.breast_cancer_screening);

  populateCheckboxGroup('side_effects_group', opts.side_effect);
}

function populateSelect(id, items) {
  const el = document.getElementById(id);
  if (!el || !items) return;
  items.forEach(item => {
    const opt = document.createElement('option');
    opt.value = item.code;
    opt.textContent = `${item.code} – ${item.label}`;
    el.appendChild(opt);
  });
}

function populateCheckboxGroup(containerId, items) {
  const container = document.getElementById(containerId);
  if (!container || !items) return;
  container.innerHTML = '';
  items.forEach(item => {
    const div = document.createElement('div');
    div.className = 'form-check form-check-inline';
    div.innerHTML = `
      <input class="form-check-input side-effect-check" type="checkbox"
             id="se_${item.code}" value="${item.code}">
      <label class="form-check-label" for="se_${item.code}"
             title="${item.description || item.label}">${item.code}</label>`;
    container.appendChild(div);
  });
}

function setupSkipLogic() {
  // New User / Revisit mutual exclusion
  const newUser = document.getElementById('is_new_user');
  const revisit = document.getElementById('is_revisit');
  const prevMethodGroup = document.getElementById('previous_method_group');

  function toggleRevisitFields() {
    if (revisit && revisit.checked) {
      prevMethodGroup?.classList.remove('skip-hidden');
    } else {
      prevMethodGroup?.classList.add('skip-hidden');
      const pm = document.getElementById('previous_method');
      if (pm) pm.value = '';
    }
  }

  if (newUser) newUser.addEventListener('change', () => {
    if (newUser.checked && revisit) revisit.checked = false;
    toggleRevisitFields();
  });
  if (revisit) revisit.addEventListener('change', () => {
    if (revisit.checked && newUser) newUser.checked = false;
    toggleRevisitFields();
  });
  toggleRevisitFields();

  // Switching method
  const switching = document.getElementById('is_switching');
  const switchReasonGroup = document.getElementById('switching_reason_group');
  function toggleSwitching() {
    if (switching && switching.checked) {
      switchReasonGroup?.classList.remove('skip-hidden');
    } else {
      switchReasonGroup?.classList.add('skip-hidden');
      const sr = document.getElementById('switching_reason');
      if (sr) sr.value = '';
    }
  }
  if (switching) switching.addEventListener('change', toggleSwitching);
  toggleSwitching();

  // Cervical cancer treatment only if status is positive (code "2")
  const ccStatus = document.getElementById('cervical_cancer_status');
  const ccTreatmentGroup = document.getElementById('cc_treatment_group');
  function toggleCCTreatment() {
    if (ccStatus && ccStatus.value === '2') {
      ccTreatmentGroup?.classList.remove('skip-hidden');
    } else {
      ccTreatmentGroup?.classList.add('skip-hidden');
      const ct = document.getElementById('cervical_cancer_treatment');
      if (ct) ct.value = '';
    }
  }
  if (ccStatus) ccStatus.addEventListener('change', toggleCCTreatment);
  toggleCCTreatment();

  // LARC removal fields only show if implant/IUD removal reason is set
  const implantReason = document.getElementById('implant_removal_reason');
  const implantTimingGroup = document.getElementById('implant_timing_group');
  function toggleImplantTiming() {
    if (implantReason && implantReason.value) {
      implantTimingGroup?.classList.remove('skip-hidden');
    } else {
      implantTimingGroup?.classList.add('skip-hidden');
    }
  }
  if (implantReason) implantReason.addEventListener('change', toggleImplantTiming);
  toggleImplantTiming();

  const iudReason = document.getElementById('iud_removal_reason');
  const iudTimingGroup = document.getElementById('iud_timing_group');
  function toggleIUDTiming() {
    if (iudReason && iudReason.value) {
      iudTimingGroup?.classList.remove('skip-hidden');
    } else {
      iudTimingGroup?.classList.add('skip-hidden');
    }
  }
  if (iudReason) iudReason.addEventListener('change', toggleIUDTiming);
  toggleIUDTiming();

  // Sex-specific fields: vasectomy only for M, tubal ligation only for F
  const sex = document.getElementById('sex');
  function toggleSexFields() {
    const val = sex ? sex.value : '';
    const vasGroup = document.getElementById('vasectomy_group');
    const tlGroup = document.getElementById('tubal_ligation_group');
    const cervicalGroup = document.getElementById('cervical_screening_section');
    const breastGroup = document.getElementById('breast_screening_section');
    const postPregSection = document.getElementById('post_preg_section');
    const larcSection = document.getElementById('larc_section');

    const pillsGroup = document.getElementById('pills_group');
    const condomsGroup = document.getElementById('condoms_group');
    const injectablesGroup = document.getElementById('injectables_group');
    const implantsGroup = document.getElementById('implants_group');
    const iudsGroup = document.getElementById('iuds_group');
    const permanentGroup = document.getElementById('permanent_group');
    const famGroup = document.getElementById('fam_group');

    if (val === 'M') {
      vasGroup?.classList.remove('skip-hidden');
      tlGroup?.classList.add('skip-hidden');
      cervicalGroup?.classList.add('skip-hidden');
      breastGroup?.classList.add('skip-hidden');

      // For males: only show condoms + vasectomy; hide all other method groups and female-only sections
      pillsGroup?.classList.add('skip-hidden');
      injectablesGroup?.classList.add('skip-hidden');
      implantsGroup?.classList.add('skip-hidden');
      iudsGroup?.classList.add('skip-hidden');
      famGroup?.classList.add('skip-hidden');
      postPregSection?.classList.add('skip-hidden');
      larcSection?.classList.add('skip-hidden');

      condomsGroup?.classList.remove('skip-hidden');
      permanentGroup?.classList.remove('skip-hidden');

      // Clear values for hidden female-only fields
      ['pills_coc_cycles', 'pills_pop_cycles', 'pills_ecp_pieces',
       'condoms_female_units',
       'injectable_dmpa_im_doses', 'injectable_dmpa_sc_pa_doses', 'injectable_dmpa_sc_si_doses'
      ].forEach(id => {
        const el = document.getElementById(id);
        if (el) el.value = '0';
      });
      ['implant_3_years', 'implant_5_years',
       'iud_copper_t', 'iud_hormonal_3_years', 'iud_hormonal_5_years',
       'fam_standard_days', 'fam_lam', 'fam_two_day'
      ].forEach(id => {
        const el = document.getElementById(id);
        if (el) el.checked = false;
      });
      ['postpartum_fp_timing', 'post_abortion_fp_timing',
       'implant_removal_reason', 'implant_removal_timing',
       'iud_removal_reason', 'iud_removal_timing'
      ].forEach(id => {
        const el = document.getElementById(id);
        if (el) el.value = '';
      });
    } else if (val === 'F') {
      vasGroup?.classList.add('skip-hidden');
      tlGroup?.classList.remove('skip-hidden');
      cervicalGroup?.classList.remove('skip-hidden');
      breastGroup?.classList.remove('skip-hidden');
      postPregSection?.classList.remove('skip-hidden');
      larcSection?.classList.remove('skip-hidden');

      pillsGroup?.classList.remove('skip-hidden');
      condomsGroup?.classList.remove('skip-hidden');
      injectablesGroup?.classList.remove('skip-hidden');
      implantsGroup?.classList.remove('skip-hidden');
      iudsGroup?.classList.remove('skip-hidden');
      permanentGroup?.classList.remove('skip-hidden');
      famGroup?.classList.remove('skip-hidden');
    } else {
      vasGroup?.classList.remove('skip-hidden');
      tlGroup?.classList.remove('skip-hidden');
      cervicalGroup?.classList.remove('skip-hidden');
      breastGroup?.classList.remove('skip-hidden');
      postPregSection?.classList.remove('skip-hidden');
      larcSection?.classList.remove('skip-hidden');

      pillsGroup?.classList.remove('skip-hidden');
      condomsGroup?.classList.remove('skip-hidden');
      injectablesGroup?.classList.remove('skip-hidden');
      implantsGroup?.classList.remove('skip-hidden');
      iudsGroup?.classList.remove('skip-hidden');
      permanentGroup?.classList.remove('skip-hidden');
      famGroup?.classList.remove('skip-hidden');
    }

    if (typeof updateMethodGroups === 'function') {
      updateMethodGroups();
    }
  }
  if (sex) sex.addEventListener('change', toggleSexFields);
  toggleSexFields();

  // Contraceptive method skip logic: one method group at a time,
  // and within injectables/implants/IUDs only one option.
  const pillIds = ['pills_coc_cycles', 'pills_pop_cycles', 'pills_ecp_pieces'];
  const condomIds = ['condoms_male_units', 'condoms_female_units'];
  const injectableIds = ['injectable_dmpa_im_doses', 'injectable_dmpa_sc_pa_doses', 'injectable_dmpa_sc_si_doses'];
  const implantIds = ['implant_3_years', 'implant_5_years'];
  const iudIds = ['iud_copper_t', 'iud_hormonal_3_years', 'iud_hormonal_5_years'];
  const permanentIds = ['tubal_ligation', 'vasectomy'];
  const famIds = ['fam_standard_days', 'fam_lam', 'fam_two_day'];

  function getInt(id) {
    const el = document.getElementById(id);
    if (!el) return 0;
    const v = parseInt(el.value, 10);
    return isNaN(v) ? 0 : v;
  }
  function getCheck(id) {
    const el = document.getElementById(id);
    return !!(el && el.checked);
  }
  function setDisabled(ids, disabled) {
    ids.forEach(id => {
      const el = document.getElementById(id);
      if (!el) return;
      el.disabled = disabled;
      if (disabled) el.classList.add('bg-light');
      else el.classList.remove('bg-light');
    });
  }

  function clampNumberInput(id, maxVal) {
    const el = document.getElementById(id);
    if (!el) return;
    const v = parseInt(el.value, 10);
    const next = isNaN(v) ? 0 : v;
    if (next < 0) el.value = '0';
    else if (typeof maxVal === 'number' && next > maxVal) el.value = String(maxVal);
  }

  function enforceMaxes() {
    clampNumberInput('pills_coc_cycles', 3);
    clampNumberInput('pills_pop_cycles', 3);
    clampNumberInput('pills_ecp_pieces', 1);
    clampNumberInput('condoms_male_units', 144);
    // female condoms: no max
    clampNumberInput('injectable_dmpa_im_doses', 1);
    clampNumberInput('injectable_dmpa_sc_pa_doses', 4);
    clampNumberInput('injectable_dmpa_sc_si_doses', 4);
  }

  function updateMethodGroups() {
    enforceMaxes();
    const pillsVal = pillIds.some(id => getInt(id) > 0);
    const condomsVal = condomIds.some(id => getInt(id) > 0);
    const injIm = getInt('injectable_dmpa_im_doses');
    const injPa = getInt('injectable_dmpa_sc_pa_doses');
    const injSi = getInt('injectable_dmpa_sc_si_doses');
    const injectablesVal = injIm > 0 || injPa > 0 || injSi > 0;
    const implantVal = implantIds.some(id => getCheck(id));
    const iudVal = iudIds.some(id => getCheck(id));
    const permanentVal = permanentIds.some(id => getCheck(id));
    const famVal = famIds.some(id => getCheck(id));

    const anyMethod = pillsVal || condomsVal || injectablesVal || implantVal || iudVal || permanentVal || famVal;

    // One method group at a time
    setDisabled(pillIds, anyMethod && !pillsVal);
    setDisabled(condomIds, anyMethod && !condomsVal);
    setDisabled(injectableIds, anyMethod && !injectablesVal);
    setDisabled(implantIds, anyMethod && !implantVal);
    setDisabled(iudIds, anyMethod && !iudVal);
    setDisabled(permanentIds, anyMethod && !permanentVal);
    setDisabled(famIds, anyMethod && !famVal);

    // Within injectables: only one type may have a non-zero dose
    if (injectablesVal) {
      setDisabled(['injectable_dmpa_im_doses'], injectablesVal && injIm === 0);
      setDisabled(['injectable_dmpa_sc_pa_doses'], injectablesVal && injPa === 0);
      setDisabled(['injectable_dmpa_sc_si_doses'], injectablesVal && injSi === 0);
    } else {
      setDisabled(injectableIds, false);
    }

    // Side effects visibility and filtering: only show when applicable and
    // show method-specific codes based on selected method type.
    const sideEffectsSection = document.getElementById('side_effects_group')?.closest('.form-section');
    const sexVal = document.getElementById('sex')?.value || '';
    const showSideEffects =
      sexVal === 'F' && (pillsVal || injectablesVal || implantVal || iudVal);

    if (sideEffectsSection) {
      if (!showSideEffects) {
        sideEffectsSection.classList.add('skip-hidden');
        document.querySelectorAll('.side-effect-check').forEach(cb => { cb.checked = false; cb.disabled = false; });
      } else {
        sideEffectsSection.classList.remove('skip-hidden');

        const allowed = new Set(['IB','HB','NB','NV','HE','MO','BR','WG','SP','AC','ETC']);
        if (iudVal) { allowed.add('CR'); allowed.add('SAP'); }
        if (injectablesVal) allowed.add('ISR');
        if (implantVal) allowed.add('INSR');

        document.querySelectorAll('.side-effect-check').forEach(cb => {
          const code = cb.value;
          const wrapper = cb.closest('.form-check');
          const ok = allowed.has(code);
          cb.disabled = !ok;
          if (!ok) cb.checked = false;
          if (wrapper) wrapper.style.display = ok ? 'inline-flex' : 'none';
        });
      }
    }
  }

  // Implants, IUDs, permanent methods, FAM: allow only one checkbox per group
  function enforceSingleCheckbox(ids, changedId) {
    ids.forEach(id => {
      if (id === changedId) return;
      const el = document.getElementById(id);
      if (el) el.checked = false;
    });
  }

  injectableIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', updateMethodGroups);
  });
  pillIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', updateMethodGroups);
  });
  condomIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', updateMethodGroups);
  });

  implantIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('change', () => {
      if (el.checked) enforceSingleCheckbox(implantIds, id);
      updateMethodGroups();
    });
  });
  iudIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('change', () => {
      if (el.checked) enforceSingleCheckbox(iudIds, id);
      updateMethodGroups();
    });
  });
  permanentIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('change', () => {
      if (el.checked) enforceSingleCheckbox(permanentIds, id);
      updateMethodGroups();
    });
  });
  famIds.forEach(id => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('change', () => {
      if (el.checked) enforceSingleCheckbox(famIds, id);
      updateMethodGroups();
    });
  });

  updateMethodGroups();
}

function collectFormData() {
  const getVal = id => (document.getElementById(id)?.value || '').trim();
  const getInt = id => parseInt(document.getElementById(id)?.value) || 0;
  const getCheck = id => document.getElementById(id)?.checked || false;
  const getBool = id => {
    const el = document.getElementById(id);
    if (!el) return null;
    if (el.type === 'checkbox') return el.checked;
    if (el.value === 'true') return true;
    if (el.value === 'false') return false;
    return null;
  };

  const sideEffects = [];
  document.querySelectorAll('.side-effect-check:checked').forEach(cb => {
    sideEffects.push(cb.value);
  });

  return {
    visit_date: getVal('visit_date'),
    is_visitor: getCheck('is_visitor'),
    nin: getVal('nin'),
    surname: getVal('surname'),
    given_name: getVal('given_name'),
    phone_number: getVal('phone_number'),
    village: getVal('village'),
    parish: getVal('parish'),
    subcounty: getVal('subcounty'),
    district: getVal('district'),
    sex: getVal('sex'),
    age: getInt('age'),
    is_new_user: getCheck('is_new_user'),
    is_revisit: getCheck('is_revisit'),
    previous_method: getVal('previous_method'),
    hts_code: getVal('hts_code'),
    counseling_individual: getCheck('counseling_individual'),
    counseling_as_couple: getCheck('counseling_as_couple'),
    counseling_om: getCheck('counseling_om'),
    counseling_se: getCheck('counseling_se'),
    counseling_wd: getCheck('counseling_wd'),
    counseling_ms: getCheck('counseling_ms'),
    is_switching: getCheck('is_switching'),
    switching_reason: getVal('switching_reason'),
    pills_coc_cycles: getInt('pills_coc_cycles'),
    pills_pop_cycles: getInt('pills_pop_cycles'),
    pills_ecp_pieces: getInt('pills_ecp_pieces'),
    condoms_male_units: getInt('condoms_male_units'),
    condoms_female_units: getInt('condoms_female_units'),
    injectable_dmpa_im_doses: getInt('injectable_dmpa_im_doses'),
    injectable_dmpa_sc_pa_doses: getInt('injectable_dmpa_sc_pa_doses'),
    injectable_dmpa_sc_si_doses: getInt('injectable_dmpa_sc_si_doses'),
    implant_3_years: getCheck('implant_3_years'),
    implant_5_years: getCheck('implant_5_years'),
    iud_copper_t: getCheck('iud_copper_t'),
    iud_hormonal_3_years: getCheck('iud_hormonal_3_years'),
    iud_hormonal_5_years: getCheck('iud_hormonal_5_years'),
    tubal_ligation: getCheck('tubal_ligation'),
    vasectomy: getCheck('vasectomy'),
    fam_standard_days: getCheck('fam_standard_days'),
    fam_lam: getCheck('fam_lam'),
    fam_two_day: getCheck('fam_two_day'),
    postpartum_fp_timing: getVal('postpartum_fp_timing'),
    post_abortion_fp_timing: getVal('post_abortion_fp_timing'),
    implant_removal_reason: getVal('implant_removal_reason'),
    implant_removal_timing: getVal('implant_removal_timing'),
    iud_removal_reason: getVal('iud_removal_reason'),
    iud_removal_timing: getVal('iud_removal_timing'),
    side_effects: sideEffects.join(','),
    cervical_screening_method: getVal('cervical_screening_method'),
    cervical_cancer_status: getVal('cervical_cancer_status'),
    cervical_cancer_treatment: getVal('cervical_cancer_treatment'),
    breast_cancer_screening: getVal('breast_cancer_screening'),
    screened_for_sti: getBool('screened_for_sti'),
    referral_number: getVal('referral_number'),
    referral_reason: getVal('referral_reason'),
    remarks: getVal('remarks'),
  };
}

function setupFormSubmission() {
  const form = document.getElementById('registrationForm');
  if (!form) return;

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.getElementById('submitBtn');
    btn.disabled = true;
    btn.innerHTML = '<span class="spinner-border spinner-border-sm me-1"></span>Saving...';

    clearValidationErrors();

    try {
      const payload = collectFormData();
      const data = await App.api('POST', '/registrations', payload);
      App.showToast('Registration saved successfully! Client #: ' + (data.data.client_number || 'Visitor'), 'success');
      form.reset();
      setDefaults();
      setupSkipLogic();
    } catch (err) {
      if (err.errors && err.errors.length > 0) {
        showValidationErrors(err.errors);
        App.showToast('Please fix the validation errors', 'error');
      } else {
        App.showToast(err.message || 'Failed to save registration', 'error');
      }
    } finally {
      btn.disabled = false;
      btn.innerHTML = '<i class="bi bi-check-circle me-1"></i>Save Registration';
    }
  });
}

function showValidationErrors(errors) {
  errors.forEach(err => {
    if (err.field) {
      const el = document.getElementById(err.field);
      if (el) {
        el.classList.add('is-invalid');
        let feedback = el.parentElement.querySelector('.invalid-feedback');
        if (!feedback) {
          feedback = document.createElement('div');
          feedback.className = 'invalid-feedback';
          el.parentElement.appendChild(feedback);
        }
        feedback.textContent = err.message;
      }
    }
  });
}

function clearValidationErrors() {
  document.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
  document.querySelectorAll('.invalid-feedback').forEach(el => el.remove());
}
