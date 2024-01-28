import '../css/FriendTable.css'

function FriendTable({ friends, filterText }) {
    const rows = [];

    friends.forEach((friend) => {
      if (
        friend.Name.toLowerCase().indexOf(
          filterText.toLowerCase()
        ) === -1
      ) {
        return;
      }
      rows.push(
        <FriendRow
          friend={friend}
          key={friend.name} />
      );
    });

    return (
      <table className="friend-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Last Contacted</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>{rows}</tbody>
      </table>
    );
}

function FriendRow({ friend }) {
    const name = friend.Name

    return (
      <tr>
        <td>{name}</td>
        <td>{friend.LastContacted}</td>
        <td>{friend.Notes}</td>
      </tr>
    );
  }

export default FriendTable
