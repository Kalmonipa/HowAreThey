import React from 'react';
import PropTypes from 'prop-types';

import '../css/SearchBar.css';

function SearchBar({
  filterText,
  onFilterTextChange,
}) {
  return (
    <form className="search-bar-form">
      <input
        type="text"
        value={filterText}
        placeholder="Search..."
        onChange={(e) => onFilterTextChange(e.target.value)}
        className="search-bar" />
    </form>
  );
}
SearchBar.propTypes = {
  filterText: PropTypes.string.isRequired,
  onFilterTextChange: PropTypes.func.isRequired,
};


export default SearchBar;
