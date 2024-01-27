import React, { useState, useEffect } from 'react';
import RandomFriendButton from './components/RandomFriendButton';

function FriendsList() {
  const friends = [
    {"ID":"1","Name":"Jack Reacher","LastContacted":"01/01/2020","Notes":""},
    {"ID":"2","Name":"John Wick","LastContacted":"25/01/2024","Notes":""},
    {"ID":"3","Name":"Sonic The Hedgehog","LastContacted":"01/01/2020","Notes":""},
    {"ID":"4","Name":"Hairy Maclairy","LastContacted":"01/01/2020","Notes":""},
    {"ID":"5","Name":"Post Malone","LastContacted":"01/01/2020","Notes":""},
    {"ID":"6","Name":"Avatar Aang","LastContacted":"25/01/2024","Notes":""},
  ];

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
      <RandomFriendButton />
    </div>
  );
}

export default FriendsList;
