import '../css/SearchBar.css'

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

  export default SearchBar;
