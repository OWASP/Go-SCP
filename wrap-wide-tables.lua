-- wrap-wide-tables.lua: pandoc filter used by build-books.sh.
--
-- Tables read from GitHub-flavored Markdown have no column widths, so the
-- PDF lays each column out at its natural width and a wide table runs past
-- the right margin. This filter gives wide tables explicit widths, as
-- fractions of the page proportional to each column's longest cell, which
-- makes LaTeX wrap the cells instead. Narrow tables keep natural widths.

-- A table whose longest cells add up to fewer characters than this fits on
-- an A4 line, so it is left alone.
local fits_on_a_line = 72

-- Characters added to each column for the space between cells, so a column
-- sized to its longest word still has room for it.
local cell_padding = 3

-- column_lengths returns the length of the longest cell in each column.
local function column_lengths(tbl)
  local lengths = {}
  for i = 1, #tbl.colspecs do
    lengths[i] = 0
  end
  local function measure(rows)
    for _, row in ipairs(rows) do
      for i, cell in ipairs(row.cells) do
        local length = utf8.len(pandoc.utils.stringify(cell.contents)) or 0
        if lengths[i] and length > lengths[i] then
          lengths[i] = length
        end
      end
    end
  end
  measure(tbl.head.rows)
  for _, body in ipairs(tbl.bodies) do
    measure(body.body)
  end
  for i = 1, #lengths do
    lengths[i] = lengths[i] + cell_padding
  end
  return lengths
end

function Table(tbl)
  local lengths = column_lengths(tbl)
  local total = 0
  for _, length in ipairs(lengths) do
    total = total + length
  end
  if total < fits_on_a_line then
    return nil
  end
  for i, spec in ipairs(tbl.colspecs) do
    tbl.colspecs[i] = { spec[1], lengths[i] / total }
  end
  return tbl
end
