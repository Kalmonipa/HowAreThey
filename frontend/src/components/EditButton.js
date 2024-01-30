import React from 'react';
import PropTypes from 'prop-types';

import '../css/EditButton.css';

const EditButton = ({ isEditable, onClick }) => {
  return (
    <button onClick={onClick} className="edit-button">
      {isEditable ? 'Save' : 'Edit Table'}
    </button>
  );
};

EditButton.propTypes = {
  isEditable: PropTypes.bool.isRequired,
  onClick: PropTypes.func.isRequired,
};

export default EditButton;
