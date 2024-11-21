/*
Copyright 2020 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sql_parser

import (
	"fmt"
	"math/rand"

	"github.com/usalko/sent/internal/sql_parser/ast"
)

// This file is used to generate random expressions to be used for testing

func newGenerator(seed int64, maxDepth int) *generator {
	g := generator{
		seed:     seed,
		r:        rand.New(rand.NewSource(seed)),
		maxDepth: maxDepth,
	}
	return &g
}

type generator struct {
	seed     int64
	r        *rand.Rand
	depth    int
	maxDepth int
}

// enter should be called whenever we are producing an intermediate node. it should be followed by a `defer g.exit()`
func (g *generator) enter() {
	g.depth++
}

// exit should be called when exiting an intermediate node
func (g *generator) exit() {
	g.depth--
}

// atMaxDepth returns true if we have reached the maximum allowed depth or the expression tree
func (g *generator) atMaxDepth() bool {
	return g.depth >= g.maxDepth
}

/*
	 Creates a random expression. It builds an expression tree using the following constructs:
	    - true/false
	    - AND/OR/NOT
	    - string literalrs, numeric literals (-/+ 1000)
	    - =, >, <, >=, <=, <=>, !=
		- &, |, ^, +, -, *, /, div, %, <<, >>
	    - IN, BETWEEN and CASE
		- IS NULL, IS NOT NULL, IS TRUE, IS NOT TRUE, IS FALSE, IS NOT FALSE

Note: It's important to update this method so that it produces all expressions that need precedence checking.
It's currently missing function calls and string operators
*/
func (g *generator) expression() ast.Expr {
	if g.randomBool() {
		return g.booleanExpr()
	}
	options := []exprF{
		func() ast.Expr { return g.intExpr() },
		func() ast.Expr { return g.stringExpr() },
		func() ast.Expr { return g.booleanExpr() },
	}

	return g.randomOf(options)
}

func (g *generator) booleanExpr() ast.Expr {
	if g.atMaxDepth() {
		return g.booleanLiteral()
	}

	options := []exprF{
		func() ast.Expr { return g.andExpr() },
		func() ast.Expr { return g.xorExpr() },
		func() ast.Expr { return g.orExpr() },
		func() ast.Expr { return g.comparison(g.intExpr) },
		func() ast.Expr { return g.comparison(g.stringExpr) },
		//func() ast.Expr { return g.comparison(g.booleanExpr) }, // this is not accepted by the parser
		func() ast.Expr { return g.inExpr() },
		func() ast.Expr { return g.between() },
		func() ast.Expr { return g.isExpr() },
		func() ast.Expr { return g.notExpr() },
		func() ast.Expr { return g.likeExpr() },
	}

	return g.randomOf(options)
}

func (g *generator) intExpr() ast.Expr {
	if g.atMaxDepth() {
		return g.intLiteral()
	}

	options := []exprF{
		func() ast.Expr { return g.arithmetic() },
		func() ast.Expr { return g.intLiteral() },
		func() ast.Expr { return g.caseExpr(g.intExpr) },
	}

	return g.randomOf(options)
}

func (g *generator) booleanLiteral() ast.Expr {
	return ast.BoolVal(g.randomBool())
}

func (g *generator) randomBool() bool {
	return g.r.Float32() < 0.5
}

func (g *generator) intLiteral() ast.Expr {
	t := fmt.Sprintf("%d", g.r.Intn(1000)-g.r.Intn((1000)))

	return ast.NewIntLiteral(t)
}

var words = []string{"ox", "ant", "ape", "asp", "bat", "bee", "boa", "bug", "cat", "cod", "cow", "cub", "doe", "dog", "eel", "eft", "elf", "elk", "emu", "ewe", "fly", "fox", "gar", "gnu", "hen", "hog", "imp", "jay", "kid", "kit", "koi", "lab", "man", "owl", "pig", "pug", "pup", "ram", "rat", "ray", "yak", "bass", "bear", "bird", "boar", "buck", "bull", "calf", "chow", "clam", "colt", "crab", "crow", "dane", "deer", "dodo", "dory", "dove", "drum", "duck", "fawn", "fish", "flea", "foal", "fowl", "frog", "gnat", "goat", "grub", "gull", "hare", "hawk", "ibex", "joey", "kite", "kiwi", "lamb", "lark", "lion", "loon", "lynx", "mako", "mink", "mite", "mole", "moth", "mule", "mutt", "newt", "orca", "oryx", "pika", "pony", "puma", "seal", "shad", "slug", "sole", "stag", "stud", "swan", "tahr", "teal", "tick", "toad", "tuna", "wasp", "wolf", "worm", "wren", "yeti", "adder", "akita", "alien", "aphid", "bison", "boxer", "bream", "bunny", "burro", "camel", "chimp", "civet", "cobra", "coral", "corgi", "crane", "dingo", "drake", "eagle", "egret", "filly", "finch", "gator", "gecko", "ghost", "ghoul", "goose", "guppy", "heron", "hippo", "horse", "hound", "husky", "hyena", "koala", "krill", "leech", "lemur", "liger", "llama", "louse", "macaw", "midge", "molly", "moose", "moray", "mouse", "panda", "perch", "prawn", "quail", "racer", "raven", "rhino", "robin", "satyr", "shark", "sheep", "shrew", "skink", "skunk", "sloth", "snail", "snake", "snipe", "squid", "stork", "swift", "swine", "tapir", "tetra", "tiger", "troll", "trout", "viper", "wahoo", "whale", "zebra", "alpaca", "amoeba", "baboon", "badger", "beagle", "bedbug", "beetle", "bengal", "bobcat", "caiman", "cattle", "cicada", "collie", "condor", "cougar", "coyote", "dassie", "donkey", "dragon", "earwig", "falcon", "feline", "ferret", "gannet", "gibbon", "glider", "goblin", "gopher", "grouse", "guinea", "hermit", "hornet", "iguana", "impala", "insect", "jackal", "jaguar", "jennet", "kitten", "kodiak", "lizard", "locust", "maggot", "magpie", "mammal", "mantis", "marlin", "marmot", "marten", "martin", "mayfly", "minnow", "monkey", "mullet", "muskox", "ocelot", "oriole", "osprey", "oyster", "parrot", "pigeon", "piglet", "poodle", "possum", "python", "quagga", "rabbit", "raptor", "rodent", "roughy", "salmon", "sawfly", "serval", "shiner", "shrimp", "spider", "sponge", "tarpon", "thrush", "tomcat", "toucan", "turkey", "turtle", "urchin", "vervet", "walrus", "weasel", "weevil", "wombat", "anchovy", "anemone", "bluejay", "buffalo", "bulldog", "buzzard", "caribou", "catfish", "chamois", "cheetah", "chicken", "chigger", "cowbird", "crappie", "crawdad", "cricket", "dogfish", "dolphin", "firefly", "garfish", "gazelle", "gelding", "giraffe", "gobbler", "gorilla", "goshawk", "grackle", "griffon", "grizzly", "grouper", "haddock", "hagfish", "halibut", "hamster", "herring", "jackass", "javelin", "jawfish", "jaybird", "katydid", "ladybug", "lamprey", "lemming", "leopard", "lioness", "lobster", "macaque", "mallard", "mammoth", "manatee", "mastiff", "meerkat", "mollusk", "monarch", "mongrel", "monitor", "monster", "mudfish", "muskrat", "mustang", "narwhal", "oarfish", "octopus", "opossum", "ostrich", "panther", "peacock", "pegasus", "pelican", "penguin", "phoenix", "piranha", "polecat", "primate", "quetzal", "raccoon", "rattler", "redbird", "redfish", "reptile", "rooster", "sawfish", "sculpin", "seagull", "skylark", "snapper", "spaniel", "sparrow", "sunbeam", "sunbird", "sunfish", "tadpole", "termite", "terrier", "unicorn", "vulture", "wallaby", "walleye", "warthog", "whippet", "wildcat", "aardvark", "airedale", "albacore", "anteater", "antelope", "arachnid", "barnacle", "basilisk", "blowfish", "bluebird", "bluegill", "bonefish", "bullfrog", "cardinal", "chipmunk", "cockatoo", "crayfish", "dinosaur", "doberman", "duckling", "elephant", "escargot", "flamingo", "flounder", "foxhound", "glowworm", "goldfish", "grubworm", "hedgehog", "honeybee", "hookworm", "humpback", "kangaroo", "killdeer", "kingfish", "labrador", "lacewing", "ladybird", "lionfish", "longhorn", "mackerel", "malamute", "marmoset", "mastodon", "moccasin", "mongoose", "monkfish", "mosquito", "pangolin", "parakeet", "pheasant", "pipefish", "platypus", "polliwog", "porpoise", "reindeer", "ringtail", "sailfish", "scorpion", "seahorse", "seasnail", "sheepdog", "shepherd", "silkworm", "squirrel", "stallion", "starfish", "starling", "stingray", "stinkbug", "sturgeon", "terrapin", "titmouse", "tortoise", "treefrog", "werewolf", "woodcock"}

func (g *generator) stringLiteral() ast.Expr {
	return ast.NewStrLiteral(g.randomOfS(words))
}

func (g *generator) stringExpr() ast.Expr {
	if g.atMaxDepth() {
		return g.stringLiteral()
	}

	options := []exprF{
		func() ast.Expr { return g.stringLiteral() },
		func() ast.Expr { return g.caseExpr(g.stringExpr) },
	}

	return g.randomOf(options)
}

func (g *generator) likeExpr() ast.Expr {
	g.enter()
	defer g.exit()
	return &ast.ComparisonExpr{
		Operator: ast.LikeOp,
		Left:     g.stringExpr(),
		Right:    g.stringExpr(),
	}
}

var comparisonOps = []ast.ComparisonExprOperator{ast.EqualOp, ast.LessThanOp, ast.GreaterThanOp, ast.LessEqualOp, ast.GreaterEqualOp, ast.NotEqualOp, ast.NullSafeEqualOp}

func (g *generator) comparison(f func() ast.Expr) ast.Expr {
	g.enter()
	defer g.exit()

	cmp := &ast.ComparisonExpr{
		Operator: comparisonOps[g.r.Intn(len(comparisonOps))],
		Left:     f(),
		Right:    f(),
	}
	return cmp
}

func (g *generator) caseExpr(valueF func() ast.Expr) ast.Expr {
	g.enter()
	defer g.exit()

	var exp ast.Expr
	var elseExpr ast.Expr
	if g.randomBool() {
		exp = valueF()
	}
	if g.randomBool() {
		elseExpr = valueF()
	}

	size := g.r.Intn(5) + 2
	var whens []*ast.When
	for i := 0; i < size; i++ {
		var cond ast.Expr
		if exp == nil {
			cond = g.booleanExpr()
		} else {
			cond = g.expression()
		}

		whens = append(whens, &ast.When{
			Cond: cond,
			Val:  g.expression(),
		})
	}

	return &ast.CaseExpr{
		Expr:  exp,
		Whens: whens,
		Else:  elseExpr,
	}
}

var arithmeticOps = []ast.BinaryExprOperator{ast.BitAndOp, ast.BitOrOp, ast.BitXorOp, ast.PlusOp, ast.MinusOp, ast.MultOp, ast.DivOp, ast.IntDivOp, ast.ModOp, ast.ShiftRightOp, ast.ShiftLeftOp}

func (g *generator) arithmetic() ast.Expr {
	g.enter()
	defer g.exit()

	op := arithmeticOps[g.r.Intn(len(arithmeticOps))]

	return &ast.BinaryExpr{
		Operator: op,
		Left:     g.intExpr(),
		Right:    g.intExpr(),
	}
}

type exprF func() ast.Expr

func (g *generator) randomOf(options []exprF) ast.Expr {
	return options[g.r.Intn(len(options))]()
}

func (g *generator) randomOfS(options []string) string {
	return options[g.r.Intn(len(options))]
}

func (g *generator) andExpr() ast.Expr {
	g.enter()
	defer g.exit()
	return &ast.AndExpr{
		Left:  g.booleanExpr(),
		Right: g.booleanExpr(),
	}
}

func (g *generator) orExpr() ast.Expr {
	g.enter()
	defer g.exit()
	return &ast.OrExpr{
		Left:  g.booleanExpr(),
		Right: g.booleanExpr(),
	}
}

func (g *generator) xorExpr() ast.Expr {
	g.enter()
	defer g.exit()
	return &ast.XorExpr{
		Left:  g.booleanExpr(),
		Right: g.booleanExpr(),
	}
}

func (g *generator) notExpr() ast.Expr {
	g.enter()
	defer g.exit()
	return &ast.NotExpr{g.booleanExpr()}
}

func (g *generator) inExpr() ast.Expr {
	g.enter()
	defer g.exit()

	expr := g.intExpr()
	size := g.r.Intn(5) + 2
	tuples := ast.ValTuple{}
	for i := 0; i < size; i++ {
		tuples = append(tuples, g.intExpr())
	}
	op := ast.InOp
	if g.randomBool() {
		op = ast.NotInOp
	}

	return &ast.ComparisonExpr{
		Operator: op,
		Left:     expr,
		Right:    tuples,
	}
}

func (g *generator) between() ast.Expr {
	g.enter()
	defer g.exit()

	var IsBetween bool
	if g.randomBool() {
		IsBetween = true
	} else {
		IsBetween = false
	}

	return &ast.BetweenExpr{
		IsBetween: IsBetween,
		Left:      g.intExpr(),
		From:      g.intExpr(),
		To:        g.intExpr(),
	}
}

func (g *generator) isExpr() ast.Expr {
	g.enter()
	defer g.exit()

	ops := []ast.IsExprOperator{ast.IsNullOp, ast.IsNotNullOp, ast.IsTrueOp, ast.IsNotTrueOp, ast.IsFalseOp, ast.IsNotFalseOp}

	return &ast.IsExpr{
		Right: ops[g.r.Intn(len(ops))],
		Left:  g.booleanExpr(),
	}
}
