import equal from 'fast-deep-equal';
import Konva from 'konva';
import { variantRules } from '../../../rules';
import { STACK_BASE_RANK } from '../../../types/constants';
import StackDirection from '../../../types/StackDirection';
import globals from '../../globals';
import HanabiCard from '../../HanabiCard';
import LayoutChild from '../../LayoutChild';

type SyncChildrenLocation = 'deck' | 'discard' | 'playStack' | 'hand';

export function onStackDirectionsChanged(directions: readonly StackDirection[]) {
  if (!variantRules.hasReversedSuits(globals.variant)) {
    return;
  }

  const stackStringsReversed = new Map<StackDirection, string>([
    [StackDirection.Undecided, ''],
    [StackDirection.Up, ''],
    [StackDirection.Down, 'Reversed'],
    [StackDirection.Finished, 'Reversed'],
  ]);

  const stackStringsUpOrDown = new Map<StackDirection, string>([
    [StackDirection.Undecided, ''],
    [StackDirection.Up, 'Up'],
    [StackDirection.Down, 'Down'],
    [StackDirection.Finished, 'Finished'],
  ]);

  // Update the stack directions (which are only used in the "Up or Down" and "Reversed" variants)
  directions.forEach((direction, i) => {
    if (globals.stackDirections[i] === direction) {
      return;
    }

    globals.stackDirections[i] = direction;

    const suit = globals.variant.suits[i];
    let text = '';
    const isUpOrDown = variantRules.isUpOrDown(globals.variant);
    if (isUpOrDown || suit.reversed) {
      const stackStrings = isUpOrDown ? stackStringsUpOrDown : stackStringsReversed;
      if (stackStrings.get(direction) === undefined) {
        throw new Error(`Not a valid stackDirection: ${direction}`);
      }
      text = stackStrings.get(direction)!;
    }

    globals.elements.suitLabelTexts[i].fitText(text);
    if (!globals.animateFast) {
      globals.layers.UI.batchDraw();
    }

    // TODO: direction arrow should be in state,
    // or calculated from state
    globals.deck
      .filter((c) => c.state?.suitIndex === i)
      .forEach((c) => c.setDirectionArrow(i));
  });
}

export function onHandsChanged(hands: ReadonlyArray<readonly number[]>) {
  syncChildren(
    'hand',
    hands,
    (i) => globals.elements.playerHands[i] as unknown as Konva.Container,
    (card, i) => card.animateToPlayerHand(i),
  );
}

export function onDiscardStacksChanged(discardStacks: ReadonlyArray<readonly number[]>) {
  syncChildren(
    'discard',
    discardStacks,
    (i) => {
      const suit = globals.variant.suits[i];
      return globals.elements.discardStacks.get(suit)! as unknown as Konva.Container;
    },
    (card) => card.animateToDiscardPile(),
  );
}

export function onPlayStacksChanged(playStacks: ReadonlyArray<readonly number[]>) {
  const childrenChanged = syncChildren(
    'playStack',
    playStacks,
    (i) => {
      const suit = globals.variant.suits[i];
      return globals.elements.playStacks.get(suit)! as unknown as Konva.Container;
    },
    (card) => card.animateToPlayStacks(),
  );
  console.log('onPlayStacksChanged:', childrenChanged);
  for (let i = 0; i < childrenChanged.length; i++) {
    if (!childrenChanged[i]) {
      continue;
    }
    const suit = globals.variant.suits[i];
    const playStack = globals.elements.playStacks.get(suit)!;
    playStack.hideCardsUnderneathTheTopCard();
  }
}

function syncChildren(
  oldLocation: SyncChildrenLocation,
  collections: ReadonlyArray<readonly number[]>,
  getCollectionUI: (i: number) => Konva.Container,
  addToCollectionUI: (card: HanabiCard, i: number) => void,
) {
  console.log(oldLocation);
  const getCard = (order: number) => globals.deck[order];

  return collections.map((collection, i) => {
    let changed = false;
    const oldParentElement = getCollectionUI(i);
    const getCurrentSorting = () => (oldParentElement.children.toArray() as LayoutChild[])
      .map((layoutChild) => layoutChild.children[0] as unknown as HanabiCard)
      .filter((card) => card.state.rank !== STACK_BASE_RANK)
      .map((card) => card.state.order);

    let current = getCurrentSorting();

    // Remove the elements that were removed
    current
      .filter((n) => !collection.includes(n))
      .map(getCard)
      .forEach((card) => {
        changed = true;
        const realState = globals.store?.getState().visibleState.deck[card.state.order];
        if (!realState || realState.location === 'deck') {
          card.animateToDeck();
        } else {
          card.removeFromParent();
        }
      });

    // Add the elements that were added
    collection
      .filter((n) => !current.includes(n))
      .map(getCard)
      .forEach((card) => {
        changed = true;
        return addToCollectionUI(card, i);
      });

    // Reorder the elements to match the collection
    collection.forEach((order, pos) => {
      current = getCurrentSorting();
      if (current.length !== collection.length) {
        throw new Error('The UI collection is out of sync with the state.');
      }

      const layoutChild = getCard(order).parent as unknown as LayoutChild;
      let sourcePosition = current.indexOf(order);
      while (sourcePosition < pos) {
        changed = true;
        layoutChild.moveUp();
        sourcePosition += 1;
      }
      while (sourcePosition > pos) {
        changed = true;
        layoutChild.moveDown();
        sourcePosition -= 1;
      }
    });

    // Verify the final result
    current = getCurrentSorting();
    if (!equal(current, collection)) {
      throw new Error('The UI collection is out of sync with the state.');
    }

    return changed;
  });
}
