import React, { useState, useEffect } from 'react';

function FriendsList() {
  const [friends, setFriends] = useState([]);

  useEffect(() => {
    fetch('http://localhost:8022/friends')
      .then(response => {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
      })
      .then(data => {
        console.log('Data received:', data);
        setFriends(data);
      })
      .catch(error => console.error('Error:', error));
  }, []);

  return (
    <div>
      <h1>Friends List</h1>
      <ul>
        {friends.length > 0 ? (
          friends.map(friend => (
            <li key={friend.ID}>
              <strong>Name:</strong> {friend.Name} |
              <strong> Last Contacted:</strong> {friend.LastContacted}
              {friend.Notes && <><strong> Notes:</strong> {friend.Notes}</>}
            </li>
          ))
        ) : (
          <p>No friends found.</p>
        )}
      </ul>
    </div>
  );
}

export default FriendsList;
