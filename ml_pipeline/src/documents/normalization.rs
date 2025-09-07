pub fn normalize(text: String, allowed: &[char]) -> String {
    let text = remove_non_printable_chars(text, allowed);
    normalize_whitespace(text)
}

fn normalize_whitespace(text: String) -> String {
    let mut result = String::with_capacity(text.len());
    let mut chars = text.chars().peekable();
    
    while let Some(c) = chars.next() {
        if c.is_whitespace() {
            // Replace any whitespace sequence with single space
            // Except when between numbers or special cases
            let next_is_digit = chars.peek().map_or(false, |c| c.is_ascii_digit());
            let prev_is_digit = !result.is_empty() && result.chars().last().unwrap().is_ascii_digit();
            
            if !(prev_is_digit && next_is_digit) {
                result.push(' ');
            }
            
            // Skip remaining whitespace
            while chars.peek().map_or(false, |c| c.is_whitespace()) {
                chars.next();
            }
        } else {
            result.push(c);
        }
    }
    
    result.trim().to_string()
}

fn remove_non_printable_chars(text: String, allowed: &[char]) -> String {
    text.chars()
        .filter(|&c| {
            c.is_ascii_graphic() || 
            allowed.contains(&c) || 
            c.is_whitespace() ||  // Preserve all whitespace initially
            (!c.is_ascii() && !c.is_control())
        })
        .collect()
}