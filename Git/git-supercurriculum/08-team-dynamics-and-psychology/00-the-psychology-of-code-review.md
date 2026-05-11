
# 00-the-psychology-of-code-review.md

- **Purpose**: To discuss the human element of code reviews, focusing on communication, empathy, and providing constructive feedback.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: `03-collaboration-and-remotes/04-pr-and-code-review-workflows.md`

---

### Code Review is a Social Activity

We often think of code review as a technical process for catching bugs. While that is one goal, it is not the primary one.

**The primary purpose of code review is to ensure shared understanding and ownership of the code.**

It's a conversation, not an exam. It's a collaboration, not a confrontation. The technical aspects are important, but the social dynamics are what make a code review process healthy and effective—or toxic and destructive.

### The Author: Overcoming Defensiveness

Receiving criticism of your work is hard. It's easy to feel defensive, especially when you've spent days on a feature.

**Mindset for Authors:**
1.  **You are not your code.** Feedback on your code is not a personal attack. It's an attempt to improve the final product, which is a shared goal.
2.  **Embrace curiosity.** When a reviewer asks a question, assume they are trying to understand, not trying to poke holes. A question like "Why did you choose this approach instead of X?" is an invitation to share your thought process.
3.  **Be grateful for the reviewer's time.** A thorough review takes significant time and effort. Thank your reviewers for their input.
4.  **Don't merge your own PR.** Even if you have the authority, let the reviewer give the final "LGTM" (Looks Good To Me) and merge. It reinforces the collaborative nature of the process.

### The Reviewer: The Responsibility of Constructive Feedback

Being a reviewer is a position of trust and responsibility. Your goal is to help, not to show off how smart you are.

**Mindset for Reviewers:**
1.  **The "Golden Rule" of Code Review**: **Review others' code as you would have them review your own.**
2.  **Start with the positive.** Always find something good to say first. "Great test coverage here!" or "I really like how you've simplified this logic." This builds goodwill and makes the author more receptive to criticism.
3.  **Ask questions, don't make statements.**
    - **Bad**: "This is inefficient. Use a Set instead of an Array."
    - **Good**: "What do you think about using a Set here? It might improve the lookup time for large inputs."
    This phrasing is collaborative. It opens a discussion rather than issuing a command.
4.  **Distinguish between blocking and non-blocking feedback.** Use prefixes to clarify the importance of your comments.
    - **`[Blocking]` or `[Request]`**: "I think this needs to be changed before merge." (e.g., a bug, a major design flaw).
    - **`[Nitpick]` or `[Suggestion]`**: "This is a minor style point, feel free to ignore." (e.g., renaming a variable for clarity, a style guide suggestion).
    - **`[Question]`**: "I'm not sure I understand this part, could you clarify?"
5.  **Automate the small stuff.** A human reviewer should not be spending their time pointing out style guide violations (e.g., wrong indentation, missing semicolons). This is the job of a linter and an automated CI check. The review should focus on the things a machine can't check: logic, design, clarity, and maintainability.
6.  **Provide solutions, not just problems.** If you suggest a change, provide a code snippet or a clear explanation of how to implement it.

### The "Comment Sandwich"

A useful technique for delivering criticism is the "comment sandwich":
1.  **Top Bun (Positive)**: "This is a clever way to handle the edge cases."
2.  **The Meat (The Criticism)**: "[Suggestion] I wonder if we could make this function name a bit more descriptive? Right now `handleData` is a bit generic. What about `parseUserDataFromAPI`?"
3.  **Bottom Bun (Encouragement)**: "Overall, this is a great change and a big improvement."

### The Goal: Psychological Safety

A healthy code review culture fosters **psychological safety**. This is the shared belief that team members can take interpersonal risks without fear of negative consequences.

- Developers should feel safe to submit imperfect, work-in-progress code for early feedback.
- Reviewers should feel safe to ask "dumb" questions.
- Authors should feel safe to disagree with a reviewer's suggestion and have a discussion about it.

When psychological safety is high, teams learn faster, innovate more, and produce higher-quality work. When it's low, developers hide their work until it's "perfect," avoid feedback, and the code review process becomes a source of anxiety and resentment.

### Key Takeaways

- Code review is a social process aimed at shared understanding.
- **Authors**: Assume good intent, be open to feedback, and separate your identity from your code.
- **Reviewers**: Be kind, ask questions instead of making demands, and automate the nitpicks.
- Use prefixes like `[Blocking]` and `[Suggestion]` to clarify the importance of your feedback.
- The ultimate goal is to create a culture of psychological safety where collaboration can thrive.
