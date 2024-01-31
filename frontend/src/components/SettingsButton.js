import React from 'react';
import PropTypes from 'prop-types';

import '../css/SettingsButton.css';

const SettingsButton = ({ isEditable, onClick }) => {
  return (
    <button onClick={onClick} className="settings-button">
      {isEditable ? 'Save' : 'Settings'}
    </button>
  );
};

SettingsButton.propTypes = {
  isEditable: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
};

export default SettingsButton;
