import React, { PureComponent } from 'react';
import { connect } from 'react-redux';
import { Label } from '../../core/components/Label/Label';
import SimplePicker from '../../core/components/Picker/SimplePicker';
import { DashboardSearchHit, OrganizationPreferences } from 'app/types';
import { setTeamHomeDashboard, setTeamTheme, setTeamTimezone, updateTeamPreferences } from './state/actions';

export interface Props {
  preferences: OrganizationPreferences;
  starredDashboards: DashboardSearchHit[];
  setTeamHomeDashboard: typeof setTeamHomeDashboard;
  setTeamTheme: typeof setTeamTheme;
  setTeamTimezone: typeof setTeamTimezone;
  updateTeamPreferences: typeof updateTeamPreferences;
}

const themes = [{ value: '', text: 'Default' }, { value: 'dark', text: 'Dark' }, { value: 'light', text: 'Light' }];

const timezones = [
  { value: '', text: 'Default' },
  { value: 'browser', text: 'Local browser time' },
  { value: 'utc', text: 'UTC' },
];

export class TeamPreferences extends PureComponent<Props> {
  onSubmitForm = event => {
    event.preventDefault();
    this.props.updateTeamPreferences();
  };

  render() {
    const { preferences, starredDashboards, setTeamHomeDashboard, setTeamTimezone, setTeamTheme } = this.props;

    const dashboards: DashboardSearchHit[] = [
      { id: 0, title: 'Default', tags: [], type: '', uid: '', uri: '', url: '' },
      ...starredDashboards,
    ];

    return (
      <form className="section gf-form-group" onSubmit={this.onSubmitForm}>
        <h3 className="page-heading">Preferences</h3>
        <div className="gf-form">
          <span className="gf-form-label width-11">UI Theme</span>
          <SimplePicker
            defaultValue={themes.find(theme => theme.value === preferences.theme)}
            options={themes}
            getOptionValue={i => i.value}
            getOptionLabel={i => i.text}
            onSelected={theme => setTeamTheme(theme.value)}
            width={20}
          />
        </div>
        <div className="gf-form">
          <Label
            width={11}
            tooltip="Not finding dashboard you want? Star it first, then it should appear in this select box."
          >
            Home Dashboard
          </Label>
          <SimplePicker
            defaultValue={dashboards.find(dashboard => dashboard.id === preferences.homeDashboardId)}
            getOptionValue={i => i.id}
            getOptionLabel={i => i.title}
            onSelected={(dashboard: DashboardSearchHit) => setTeamHomeDashboard(dashboard.id)}
            options={dashboards}
            placeholder="Chose default dashboard"
            width={20}
          />
        </div>
        <div className="gf-form">
          <label className="gf-form-label width-11">Timezone</label>
          <SimplePicker
            defaultValue={timezones.find(timezone => timezone.value === preferences.timezone)}
            getOptionValue={i => i.value}
            getOptionLabel={i => i.text}
            onSelected={timezone => setTeamTimezone(timezone.value)}
            options={timezones}
            width={20}
          />
        </div>
        <div className="gf-form-button-row">
          <button type="submit" className="btn btn-success">
            Save
          </button>
        </div>
      </form>
    );
  }
}

function mapStateToProps(state) {
  return {
    preferences: state.team.preferences,
    starredDashboards: state.user.starredDashboards,
  };
}

const mapDispatchToProps = {
  setTeamHomeDashboard,
  setTeamTimezone,
  setTeamTheme,
  updateTeamPreferences,
};

export default connect(mapStateToProps, mapDispatchToProps)(TeamPreferences);
