-- fix-internal-links.lua: pandoc filter used by build-books.sh.
--
-- build-books.sh combines every chapter into one document with --file-scope,
-- which prefixes each heading id with the chapter's path, e.g. the heading
-- "Validation" in src/input-validation/validation.md becomes
-- "src__input-validation__validationmd__validation". Links between chapters
-- are still written as relative Markdown paths, which work on GitHub and
-- GitBook but point nowhere inside a single book. This filter rewrites them:
--
--   * a link to a chapter goes to its first heading, or to the heading named
--     by its #fragment. Paths starting with "/" are relative to the book
--     root (the book-root metadata field), as GitBook treats them;
--   * a link to any other file in the repository (such as a .go example)
--     goes to that file on GitHub, using the source-url metadata field;
--   * a link it cannot resolve fails the build, so broken links are caught
--     instead of shipped.
--
-- The Markdown sources are never modified.

-- Pure helpers ------------------------------------------------------------

-- has_scheme reports whether target is an absolute URL such as https: or
-- mailto:.
local function has_scheme(target)
  return target:match("^%a[%w+.-]*:") ~= nil
end

-- normalize resolves "." and ".." segments, e.g.
-- "src/database-security/../output-encoding/README.md" ->
-- "src/output-encoding/README.md". It returns nil if the path climbs above
-- the repository root.
local function normalize(path)
  local parts = {}
  for part in path:gmatch("[^/]+") do
    if part == ".." then
      if #parts == 0 then
        return nil
      end
      table.remove(parts)
    elseif part ~= "." then
      table.insert(parts, part)
    end
  end
  return table.concat(parts, "/")
end

-- id_prefix returns the prefix --file-scope gives to heading ids from the
-- chapter at path: lowercased, "/" replaced by "__", dots removed.
local function id_prefix(path)
  return (path:lower():gsub("/", "__"):gsub("%.", ""))
end

-- State shared by the two passes ------------------------------------------

local chapters = {} -- normalized chapter path -> id prefix
local is_prefix = {} -- id prefix -> true
for _, file in ipairs(PANDOC_STATE.input_files) do
  local path = normalize(file)
  chapters[path] = id_prefix(path)
  is_prefix[chapters[path]] = true
end

local heading_ids = {} -- every heading id in the book
local first_heading = {} -- id prefix -> id of the chapter's first heading
local source_url -- base URL for files that are not chapters
local book_root -- directory that "/" in a link refers to
local unresolved = {} -- links that could not be resolved

-- file_exists reports whether path (relative to the repository root, where
-- pandoc runs) can be opened. Directories count as existing.
local function file_exists(path)
  local f = io.open(path, "r")
  if f then
    f:close()
    return true
  end
  return false
end

-- resolve returns the new target for a link, or nil if it cannot be
-- resolved. It returns target unchanged for links that need no rewriting.
local function resolve(target)
  if has_scheme(target) then
    return target
  end

  local path, fragment = target:match("^([^#]*)#?(.*)$")
  if path == "" then
    -- A link within the book. For a link to a chapter in the same
    -- directory, pandoc has already replaced the path with the chapter's
    -- bare id prefix, which is not a heading id itself.
    if heading_ids[fragment] then
      return target
    end
    if is_prefix[fragment] and first_heading[fragment] then
      return "#" .. first_heading[fragment]
    end
    return nil
  end

  if path:sub(1, 1) == "/" then
    if not book_root then
      return nil
    end
    path = book_root .. path
  end
  path = normalize(path)
  if not path then
    return nil
  end

  local prefix = chapters[path]
  if prefix then
    if fragment == "" then
      return first_heading[prefix] and "#" .. first_heading[prefix]
    end
    local id = prefix .. "__" .. fragment
    return heading_ids[id] and "#" .. id
  end

  if file_exists(path) and source_url then
    return source_url .. path .. (fragment ~= "" and "#" .. fragment or "")
  end
  return nil
end

-- Passes --------------------------------------------------------------------

-- Pass 1 records every heading id and the source URL, since a link can
-- point at a heading that appears later in the book.
local collect = {
  Header = function(el)
    heading_ids[el.identifier] = true
    for _, prefix in pairs(chapters) do
      if not first_heading[prefix] and el.identifier:sub(1, #prefix + 2) == prefix .. "__" then
        first_heading[prefix] = el.identifier
      end
    end
  end,
  Meta = function(meta)
    if meta["source-url"] then
      source_url = pandoc.utils.stringify(meta["source-url"])
    end
    if meta["book-root"] then
      book_root = pandoc.utils.stringify(meta["book-root"])
    end
  end,
}

-- Pass 2 rewrites links and fails the build if any could not be resolved.
local rewrite = {
  Link = function(el)
    local target = resolve(el.target)
    if not target then
      table.insert(unresolved, el.target)
      return nil
    end
    el.target = target
    return el
  end,
  Pandoc = function(doc)
    for _, prefix in pairs(chapters) do
      if not first_heading[prefix] then
        error("fix-internal-links.lua: no headings found with prefix " .. prefix
          .. "; has pandoc's --file-scope id scheme changed?")
      end
    end
    if #unresolved > 0 then
      error("fix-internal-links.lua: cannot resolve " .. #unresolved .. " link(s):\n  "
        .. table.concat(unresolved, "\n  "))
    end
    return doc
  end,
}

return { collect, rewrite }
