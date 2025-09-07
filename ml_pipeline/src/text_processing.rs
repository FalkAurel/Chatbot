#[test]
pub fn extract_keywords() {
    use rake::{Rake, StopWords};

    let mut stop_words: StopWords = StopWords::new();
    stop_words.insert(".".to_string());
    stop_words.insert("Gay".to_string());
    stop_words.insert("Aids".to_string());

    let rake: Rake = Rake::new(stop_words);

    let keywords: Vec<rake::KeywordScore> = rake.run(
        "Cats continue to benefit people in a wide range of ways –
        providing eco-friendly vermin control on farms and stables,
        companionship in family homes or cheering up the residents in
        a care home.
        Living with a feline friend brings many benefits and it also
        places legal responsibilities on the owner. Much of the law in
        relation to animals has now been consolidated into the Animal
        Welfare Act 2006 – England and Wales. Scotland and Northern
        Ireland have equivalent legislation – the Animal Health and
        Welfare (Scotland) Act 2006 and the Welfare of Animals Act
        (Northern Ireland) 2011. The Act applies to both domestic and
        feral cats and in addition to cruelty offences, it places a duty
        of care on owners and those responsible for looking after cats
        to ensure that their welfare needs are met. "
    );
    
    dbg!(keywords);
}