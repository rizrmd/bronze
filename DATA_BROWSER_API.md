# Data Browser API Documentation

The Bronze Backend now includes a unified data browser API that allows you to browse and extract data from Excel, CSV, MDB files, and **any file type as CSV** stored in S3/MinIO.

## Features

- **Excel Support** (.xlsx, .xls): Browse sheets, extract data with pagination
- **CSV Support** (.csv): Parse and browse CSV data with configurable delimiters
- **Universal CSV Support**: Treat ANY file as CSV (txt, dat, log, custom extensions)
- **MDB Support** (.mdb): Placeholder for Microsoft Access database files (future enhancement)
- **Unified Interface**: Single API endpoint for all supported file types
- **Pagination**: Built-in support for large datasets with offset/limit
- **Metadata Extraction**: Automatically extract sheet names, columns, and row counts
- **Smart Detection**: Auto-detect delimiters and headers for CSV files
- **Flexible Parsing**: Supports comma, semicolon, tab, and pipe delimiters

## API Endpoints

### 1. Browse Data from File

**Endpoint**: `POST /api/data/browse`

Browse and extract data from a specific file in S3.

#### Request Body

```json
{
  "file_name": "data.xlsx",        // Required: Name of the file to browse
  "sheet_name": "Sheet1",         // Optional: Sheet name for Excel files
  "max_rows": 100,                // Optional: Max rows to return (1-10000, default: 100)
  "offset": 0,                    // Optional: Number of rows to skip (default: 0)
  "has_headers": true             // Optional: Whether first row contains headers (default: false)
}
```

#### Response

```json
{
  "success": true,
  "message": "Excel file processed successfully",
  "data_type": "excel",
  "file_name": "data.xlsx",
  "sheet_name": "Sheet1",
  "columns": ["Name", "Age", "City"],
  "rows": [
    ["John Doe", "30", "New York"],
    ["Jane Smith", "25", "Los Angeles"]
  ],
  "total_rows": 1000,
  "row_count": 2,
  "offset": 0,
  "has_headers": true,
  "sheets": ["Sheet1", "Sheet2", "Data"]
}
```

#### Usage Examples

**Browse Excel file with first sheet:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "sales_data.xlsx",
    "has_headers": true,
    "max_rows": 50
  }'
```

**Browse specific Excel sheet:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "report.xlsx",
    "sheet_name": "Q4_Data",
    "has_headers": true,
    "max_rows": 100,
    "offset": 200
  }'
```

**Browse CSV file:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "customers.csv",
    "has_headers": true,
    "max_rows": 25
  }'
```

**Browse MDB database:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "database.mdb",
    "sheet_name": "Customers",
    "has_headers": true,
    "max_rows": 50
  }'
```

**Browse ACCDB database (newer Access):**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "new_database.accdb",
    "sheet_name": "Products",
    "has_headers": true,
    "max_rows": 100
  }'
```

**Browse ANY file as CSV (txt, dat, log, etc.):**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "server_logs.txt",
    "treat_as_csv": true,
    "auto_detect_headers": true,
    "has_headers": false,
    "max_rows": 100
  }'
```

**Stream large CSV file:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "massive_dataset.csv",
    "stream_mode": true,
    "chunk_size": 5000,
    "max_rows": 1000000,
    "has_headers": true,
    "auto_detect_headers": false
  }'
```

**Auto-detect headers and delimiters:**
```bash
curl -X POST http://localhost:8060/api/data/browse \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "unknown_format.dat",
    "treat_as_csv": true,
    "auto_detect_headers": true,
    "max_rows": 50
  }'
```

### 2. List Data Files

**Endpoint**: `GET /api/data/files`

List all supported data files (Excel, CSV, MDB) with their metadata.

#### Response

```json
{
  "success": true,
  "message": "Data files listed successfully",
  "files": [
    {
      "name": "sales_data.xlsx",
      "size": 1024000,
      "last_modified": "2023-01-01T00:00:00Z",
      "data_type": "excel",
      "sheets": ["Sheet1", "Sheet2"],
      "columns": ["Product", "Sales", "Date"],
      "row_count": 1500
    },
    {
      "name": "customers.csv",
      "size": 256000,
      "last_modified": "2023-01-02T00:00:00Z",
      "data_type": "csv",
      "columns": ["Name", "Email", "Phone"],
      "row_count": 500
    }
  ],
  "count": 2
}
```

#### Usage Example

```bash
curl -X GET http://localhost:8060/api/data/files
```

## Streaming Support

For very large files (hundreds of MB or GB), use streaming mode to avoid loading the entire file into memory.

### Streaming Request

```json
{
  "file_name": "large_dataset.csv",
  "stream_mode": true,
  "chunk_size": 5000,
  "max_rows": 100000,
  "has_headers": true,
  "auto_detect_headers": true
}
```

### Streaming Response

Streaming responses are sent as multiple JSON objects over the same HTTP connection:

1. **Initial Metadata**: File info and streaming configuration
2. **Header Information**: Column names and header detection results
3. **Data Chunks**: Groups of rows with progress information
4. **Completion Marker**: Final chunk indicating completion

```json
{"chunk_size":5000,"data_type":"csv","file_name":"large_dataset.csv","has_headers":true,"message":"Streaming CSV data","offset":0,"streaming":true,"success":true}
{"columns":["Name","Age","City"],"has_headers":true,"success":true}
{"data":[["John","30","NYC"],["Jane","25","LA"]],"progress":{"current_row":5000,"processed":2},"row_count":2,"success":true}
{"data":[["Bob","35","Chicago"],["Alice","28","Boston"]],"progress":{"current_row":10000,"processed":4},"row_count":2,"success":true}
{"complete":true,"message":"Streaming completed","row_count":100000,"total_rows":150000}
```

### Streaming Benefits

- **Memory Efficiency**: Processes files without loading entire content into memory
- **Real-time Progress**: Client receives data as soon as it's processed
- **Resumable**: Clients can track progress and resume if interrupted
- **Scalable**: Handles files much larger than available memory
- **Responsive**: First data arrives quickly, improving user experience

## File Type Support

### Excel Files (.xlsx, .xls, .xlsm)

- **Full Support**: Multiple sheets, cell formatting, data types
- **Macro Support**: XLSM files with macros are fully supported
- **Sheet Discovery**: Automatically lists all available sheets
- **Cell Parsing**: Handles numbers, dates, text, and formulas (returns calculated values)
- **Large File Handling**: Efficient processing for large workbooks

### CSV Files (.csv)

- **Full Support**: Standard CSV parsing with configurable options
- **Header Detection**: Optional header row processing
- **Encoding**: Supports UTF-8 encoded files
- **Delimiter Detection**: Auto-detects comma, semicolon, tab, or pipe delimiters
- **Streaming Support**: Stream large CSV files without memory issues

### Universal CSV Support (Any File Type)

- **Treat as CSV**: Any file can be processed as CSV regardless of extension
- **Smart Parsing**: Auto-detects delimiters and headers for non-standard files
- **Flexible Input**: Works with .txt, .dat, .log, or custom extensions
- **Mixed Data**: Handles files with mixed content types

### MDB Files (.mdb, .accdb)

- **Full Support**: Microsoft Access database files (both MDB and ACCDB)
- **Table Discovery**: Automatically lists all database tables
- **SQL Querying**: Direct database access with proper SQL queries
- **Data Types**: Handles all Access data types with proper conversion
- **Multi-Version Support**: Works with both older MDB and newer ACCDB formats
- **Pagination**: Built-in OFFSET/FETCH and fallback pagination

## Performance Considerations

1. **Memory Usage**: Non-streaming mode loads files into memory. Use streaming for files >100MB.
2. **Streaming Mode**: Processes files in chunks with configurable memory footprint.
3. **Pagination**: Use `max_rows` and `offset` parameters for large datasets.
4. **Concurrent Requests**: The API can handle multiple simultaneous streaming requests.
5. **Chunk Size**: Adjust `chunk_size` based on memory constraints and network conditions.

## Error Handling

### Common Error Responses

**File not found:**
```json
{
  "success": false,
  "message": "Failed to download file",
  "error": "file not found"
}
```

**Unsupported file type:**
```json
{
  "success": false,
  "message": "Unsupported file type"
}
```

**Invalid parameters:**
```json
{
  "success": false,
  "message": "File name is required"
}
```

**Sheet not found (Excel):**
```json
{
  "success": false,
  "message": "sheet 'InvalidSheet' not found"
}
```

## Integration with Existing Features

The data browser integrates seamlessly with existing Bronze backend features:

- **File Upload**: Use existing `/api/files` endpoint to upload data files
- **File Management**: Use existing file management endpoints to organize data
- **Job Processing**: Data files can be processed through existing job pipeline
- **Bucket Management**: Organize data files across different buckets

## Security Considerations

- **File Access**: Data browser respects existing MinIO/S3 permissions
- **Path Validation**: All file paths are validated to prevent directory traversal
- **Resource Limits**: Built-in limits prevent excessive resource consumption
- **Access Logs**: All data access is logged through existing logging infrastructure

## Future Enhancements

1. **MDB Support**: Full Microsoft Access database support
2. **Advanced Filtering**: Built-in filtering and search capabilities
3. **Export Options**: Export filtered data to different formats
4. **Caching**: Intelligent caching for frequently accessed data
5. **Streaming**: Real-time data streaming for very large datasets
6. **Authentication**: Role-based access control for sensitive data