import React, { useState, useEffect } from 'react';

function App() {
  // State to store the data
  const [data, setData] = useState([]);

  // Function to fetch data from your server
  const fetchData = async () => {
    try {
      const response = await fetch('http://0.0.0.0:8080/friends'); // Replace with your server URL
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const jsonData = await response.json();
      setData(jsonData);
    } catch (error) {
      console.error('Fetching error:', error);
    }
  };

  // useEffect to fetch data on component mount
  useEffect(() => {
    fetchData();
  }, []); // Empty dependency array means this effect runs once on mount

  // Render the data in a simple list
  return (
    <div>
      <h1>Contacts</h1>
      <ul>
        {data.map((item) => (
          <li key={item.ID}>
            Name: {item.Name}, Last Contacted: {item.LastContacted}, Notes: {item.Notes}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;
