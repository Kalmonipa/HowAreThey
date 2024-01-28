import React from "react";
import "../css/RandomFriendButton.css";

/**
 * Creates a button that can be used to select a friend from the list
 * @param {onRandomFriendSelect} onRandomFriendSelect
 * @returns
 */
function RandomFriendButton({ onRandomFriendSelect }) {
  return (
    <button onClick={onRandomFriendSelect} className="random-friend-button">
      Pick a friend to contact
    </button>
  );
}

export default RandomFriendButton;
