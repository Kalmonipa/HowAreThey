import React, { useState } from 'react';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import PropTypes from 'prop-types';

import '../css/AddFriendModal.css';

const AddFriendModal = ({ show, onClose, onSaved }) => {
  const [name, setName] = useState('');
  const [selectedDate, setSelectedDate] = useState(null);
  const [notes, setNotes] = useState('');

  const handleSave = () => {
    if (!name.trim()) {
      window.alert('Name field is empty');
      return;
    }
    if (selectedDate === null) {
      window.alert('No date selected. Enter approximate date if unknown');
      return;
    }

    const formattedDate = selectedDate
      ? selectedDate.toLocaleDateString('en-GB', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric'
      })
      : '';

    const friendData = {
      Name: name,
      LastContacted: formattedDate,
      Notes: notes
    };

    fetch(`http://localhost:8080/friends`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(friendData),
    }).then(response => {
      if (response.ok) {
        onSaved();
      }
      return response.json();
    }).then(() => {
      onClose();
    }).catch(error => console.error('Error:', error));
  };

if (!show) {
    return null;
}

return (
  <div className="add-friend-modal-backdrop">
    <div className="add-friend-modal">
      <h3 className="header">
        Add a new friend
      </h3>
      <div className="input-row">
        <input className='name-input'
          type="text"
          placeholder="Name..."
          value={name}
          onChange={e => setName(e.target.value)}
        />
        <DatePicker className='date-input'
          selected={selectedDate}
          onChange={date => setSelectedDate(date)}
          dateFormat="dd/MM/yyyy"
          placeholderText="Select a date..."
        />
      </div>
      <textarea className='notes-input'
        placeholder="Notes..."
        value={notes}
        onChange={e => setNotes(e.target.value)}
        rows="10"
        cols="30"
      />
      <div className="button-row">
        <button className='save-button' onClick={handleSave}>Save</button>
        <button className='close-button' onClick={onClose}>Close</button>
      </div>
    </div>
  </div>
);

};
  AddFriendModal.propTypes = {
  show: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSaved: PropTypes.func.isRequired,
};

export default AddFriendModal;
