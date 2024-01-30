import React from 'react';
import PropTypes from 'prop-types';

import '../css/FriendTable.css';
import '../css/EditBoxes.css';

function FriendTable({ friends, filterText, isEditable }) {

  const rows = friends.map((friend) => {
    if (friend.Name.toLowerCase().includes(filterText.toLowerCase())) {
      return <FriendRow friend={friend} key={friend.ID} editable={isEditable} />;
    }
    return null;
  });

  return (
    <div>
      <table className="friend-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Last Contacted</th>
            <th>Notes</th>
          </tr>
        </thead>
      <tbody>{rows}</tbody>
      </table>
    </div>
  );
}

FriendTable.propTypes = {
  friends: PropTypes.arrayOf(PropTypes.object).isRequired,
  filterText: PropTypes.string.isRequired,
};

function FriendRow({ friend, editable }) {
  const renderCell = (content, isEditable) => {
    return isEditable ? (
      <td>
        <input type="text" defaultValue={content} className="editable-input" />
      </td>
    ) : (
      <td>{content}</td>
    );
  };

  return (
    <tr>
      {renderCell(friend.ID, false)}
      {renderCell(friend.Name, editable)}
      {renderCell(friend.LastContacted, editable)}
      {renderCell(friend.Notes, editable)}
    </tr>
  );
}


FriendRow.propTypes = {
  friend: PropTypes.shape({
    Name: PropTypes.string,
    LastContacted: PropTypes.string,
    Notes: PropTypes.string,
  }).isRequired,
};

export default FriendTable;
